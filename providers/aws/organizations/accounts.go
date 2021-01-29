package organizations

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/organizations"
	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
)

type Account struct {
	ID              uint    `gorm:"primarykey"`
	AccountID       *string `neo:"unique"`
	Arn             *string `neo:"unique"`
	Email           *string
	JoinedMethod    *string
	JoinedTimestamp *time.Time
	name            *string
	status          *string
}

func (Account) TableName() string {
	return "aws_organizations_accounts"
}

func (c *Client) transformAccounts(values []*organizations.Account) ([]*Account, error) {
	var tValues []*Account
	for _, value := range values {
		tValues = append(tValues, &Account{
			AccountID:       value.Id,
			Arn:             value.Arn,
			Email:           value.Email,
			JoinedMethod:    value.JoinedMethod,
			JoinedTimestamp: value.JoinedTimestamp,
			name:            value.Name,
			status:          value.Status,
		})
	}
	return tValues, nil
}

var AccountTables = []interface{}{
	&Account{},
}

func (c *Client) accounts(gConfig interface{}) error {
	var config organizations.ListAccountsInput
	err := mapstructure.Decode(gConfig, &config)
	if err != nil {
		return err
	}

	// TODO: This doesn't work, since the account ids are not coming from the client but from the sdk call
	c.db.Where("account_id", c.accountID).Delete(AccountTables...)

	for {
		output, err := c.svc.ListAccounts(&config)
		if err != nil {
			return err
		}
		tValues, err := c.transformAccounts(output.Accounts)
		if err != nil {
			return err
		}
		c.db.ChunkedCreate(tValues)
		c.log.Info("Fetched resources", zap.String("resource", "organizations.accounts"), zap.Int("count", len(output.Accounts)))
		if aws.StringValue(output.NextToken) == "" {
			break
		}
		config.NextToken = output.NextToken
	}
	return nil
}
