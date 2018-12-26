package cost

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/davecgh/go-spew/spew"
)

const (
	Monthly = "MONTHLY"
	Daily   = "DAILY"
)

// Explorer -
type Explorer interface {
	GetCostAndUsage(*costexplorer.GetCostAndUsageInput) (*costexplorer.GetCostAndUsageOutput, error)
	GetCostForecast(*costexplorer.GetCostForecastInput) (*costexplorer.GetCostForecastOutput, error)
}

// AWS -
type AWS struct {
	CostExplorer Explorer
}

// AWSReportOptions -
type AWSReportOptions struct {
	Start       time.Time
	End         time.Time
	Granularity string
}

func (opt *AWSReportOptions) costAndUsageInput() *costexplorer.GetCostAndUsageInput {
	// FIX THIS HARD CODE
	metrics := []string{"BlendedCost"} // AmortizedCost, BlendedCost, NetAmortizedCost, NetUnblendedCost,
	// NormalizedUsageAmount, UnblendedCost, and UsageQuantity.

	return &costexplorer.GetCostAndUsageInput{
		Filter:      nil,
		GroupBy:     nil,
		Granularity: aws.String(opt.Granularity),
		Metrics:     aws.StringSlice(metrics),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(opt.Start.Format(opt.timeFormat())),
			End:   aws.String(opt.End.Format(opt.timeFormat())),
		},
	}
}

func (opt *AWSReportOptions) forcastInput() *costexplorer.GetCostForecastInput {
	// FIX THIS HARD CODE
	var metric = "BLENDED_COST" // forecastは大文字を要求される.
	var level int64 = 51        // forcaseの最低限の信頼度..?
	return &costexplorer.GetCostForecastInput{
		Granularity:             aws.String(opt.Granularity),
		Metric:                  aws.String(metric),
		PredictionIntervalLevel: aws.Int64(level),
		TimePeriod: &costexplorer.DateInterval{
			Start: aws.String(opt.Start.Format(opt.timeFormat())),
			End:   aws.String(opt.End.Format(opt.timeFormat())),
		},
	}
}

// Report -
func (aws *AWS) Report(opt *AWSReportOptions, f func(*costexplorer.GetCostAndUsageOutput) error) error {
	in := opt.costAndUsageInput()
	for {
		out, err := aws.CostExplorer.GetCostAndUsage(in)
		if err != nil {
			return err
		}
		if err := f(out); err != nil {
			return err
		}

		in.NextPageToken = out.NextPageToken
		if in.NextPageToken == nil {
			break
		}
	}
	return nil
}

// Forcast -
func (aws *AWS) Forecast(opt *AWSReportOptions, f func(*costexplorer.GetCostForecastOutput) error) error {
	in := opt.forcastInput()
	out, err := aws.CostExplorer.GetCostForecast(in)
	if err != nil {
		return err
	}
	return f(out)
}

// DumpReport -
func (aws *AWS) DumpReport(opt *AWSReportOptions) error {
	return aws.Report(opt, func(out *costexplorer.GetCostAndUsageOutput) error {
		spew.Dump(out)
		return nil
	})
}

// DumpForecast -
func (aws *AWS) DumpForecast(opt *AWSReportOptions) error {
	return aws.Forecast(opt, func(out *costexplorer.GetCostForecastOutput) error {
		spew.Dump(out)
		return nil
	})
}

func (opt *AWSReportOptions) timeFormat() string {
	return "2006-01-02"
}
