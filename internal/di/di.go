package di

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/ymgyt/slack/webhook"

	"github.com/ymgyt/appkit/envvar"
	"github.com/ymgyt/costman/cost"
)

// Config -
type Config struct {
	AWSAccessKeyID string `envvar:"AWS_ACCESS_KEY_ID"`
	AWSSecretKey   string `envvar:"AWS_SECRET_ACCESS_KEY"`
	AWSRegion      string `envvar:"AWS_REGION"`

	SlackWebhookURL string `envvar:"SLACK_WEBHOOK_URL"`
	SlackChannel    string `envvar:"SLACK_CHANNEL"`
	SlackUsername   string `envvar:"SLACK_USERNAME"`
	SlackIconEmoji  string `envvar:"SLACK_ICONEMOJI"`
}

// Services -
type Services struct {
	AWS     *cost.AWS
	Webhook *webhook.Client
	Config  *Config
}

// MustServices -
func MustServices() *Services {

	cfg := mustConfig()
	ce := mustAWSCostExplorer(cfg)

	return &Services{
		AWS: &cost.AWS{
			CostExplorer: ce,
		},
		Webhook: mustSlackWebhook(cfg),
		Config:  cfg,
	}
}

func mustAWSCostExplorer(cfg *Config) *costexplorer.CostExplorer {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.AWSRegion),
		Credentials: credentials.NewStaticCredentials(cfg.AWSAccessKeyID, cfg.AWSSecretKey, ""),
	})
	if err != nil {
		panic(err)
	}

	return costexplorer.New(sess)
}

func mustSlackWebhook(cfg *Config) *webhook.Client {
	wh, err := webhook.New(webhook.Config{
		URL:       cfg.SlackWebhookURL,
		Channel:   cfg.SlackChannel,
		Username:  cfg.SlackUsername,
		IconEmoji: cfg.SlackIconEmoji,
		Timeout:   0,
	})
	if err != nil {
		panic(err)
	}
	return wh
}

func mustConfig() *Config {
	cfg := &Config{}
	if err := envvar.Inject(cfg); err != nil {
		panic(err)
	}
	return cfg
}
