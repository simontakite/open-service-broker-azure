package mysqldb

import (
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/generate"
	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (a *allInOneManager) ValidateBindingParameters(
	bindingParameters service.BindingParameters,
) error {
	// There are no parameters for binding to MySQL, so there is nothing
	// to validate
	return nil
}

func (a *allInOneManager) Bind(
	instance service.Instance,
	_ service.BindingParameters,
) (service.BindingDetails, service.SecureBindingDetails, error) {
	dt, ok := instance.Details.(*allInOneMysqlInstanceDetails)
	if !ok {
		return nil, nil, fmt.Errorf(
			"error casting instance.Details as *allInOneMysqlInstanceDetails",
		)
	}
	sdt, ok := instance.SecureDetails.(*allInOneMysqlSecureInstanceDetails)
	if !ok {
		return nil, nil, fmt.Errorf(
			"error casting instance.SecureDetails as " +
				"*allInOneMysqlSecureInstanceDetails",
		)
	}

	userName := generate.NewIdentifier()
	password := generate.NewPassword()

	db, err := a.getDBConnection(dt, sdt)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close() // nolint: errcheck

	// Open doesn't open a connection. Validate DSN data:
	if err = db.Ping(); err != nil {
		return nil, nil, err
	}

	if _, err = db.Exec(
		fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", userName, password),
	); err != nil {
		return nil, nil, fmt.Errorf(
			`error creating user "%s": %s`,
			userName,
			err,
		)
	}

	if _, err = db.Exec(
		fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, "+
			"INDEX, ALTER, CREATE TEMPORARY TABLES, LOCK TABLES, "+
			"CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, "+
			"EXECUTE, REFERENCES, EVENT, "+
			"TRIGGER ON %s.* TO '%s'@'%%'",
			dt.DatabaseName, userName)); err != nil {
		return nil, nil, fmt.Errorf(
			`error granting permission to "%s": %s`,
			userName,
			err,
		)
	}

	return &mysqlBindingDetails{
			LoginName: userName,
		},
		&mysqlSecureBindingDetails{
			Password: password,
		},
		nil
}

func (a *allInOneManager) GetCredentials(
	instance service.Instance,
	binding service.Binding,
) (service.Credentials, error) {
	dt, ok := instance.Details.(*allInOneMysqlInstanceDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting instance.Details as *allInOneMysqlInstanceDetails",
		)
	}
	bd, ok := binding.Details.(*mysqlBindingDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting binding.Details as *mysqlBindingDetails",
		)
	}
	sbd, ok := binding.SecureDetails.(*mysqlSecureBindingDetails)
	if !ok {
		return nil, fmt.Errorf(
			"error casting binding.SecureDetails as *mysqlSecureBindingDetails",
		)
	}
	return &Credentials{
		Host:     dt.FullyQualifiedDomainName,
		Port:     3306,
		Database: dt.DatabaseName,
		Username: fmt.Sprintf("%s@%s", bd.LoginName, dt.ServerName),
		Password: sbd.Password,
	}, nil
}