package scenario

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"

	"github.com/ymgyt/costman/cost"
	"github.com/ymgyt/slack/webhook"
)

// Notificator -
type Notificator interface {
	SendPayload(*webhook.Payload) error
}

type Reporter interface {
	Report(*cost.AWSReportOptions, func(*costexplorer.GetCostAndUsageOutput) error) error
	Forecast(*cost.AWSReportOptions, func(*costexplorer.GetCostForecastOutput) error) error
}

// AWSDailyReport Report three things.
// - cost of previous day.
// - sum of daily cost.
// - monthly forecast.
func AWSDailyReport(reporter Reporter, notificator Notificator) error {
	ctx := &awsDailyReportContext{reporter: reporter, notificator: notificator}

	err := ctx.
		costOfYesterday().
		monthlySumOfCost().
		monthlyForcast().
		notify().
		Err

	return err
}

type awsDailyReportContext struct {
	CostOfYesterday    *costexplorer.ResultByTime
	MonthlySumOfCost   *costexplorer.ResultByTime
	MonthlyForcast     *costexplorer.GetCostForecastOutput
	Err                error
	SkipMonthlyForcast bool

	reporter    Reporter
	notificator Notificator
}

func (ctx *awsDailyReportContext) costOfYesterday() *awsDailyReportContext {
	if ctx.Err != nil {
		return ctx
	}
	opt := &cost.AWSReportOptions{
		Start:       time.Now().UTC().AddDate(0, 0, -2),
		End:         time.Now().UTC().AddDate(0, 0, -1),
		Granularity: cost.Daily,
	}
	/*
		(*costexplorer.GetCostAndUsageOutput)(0xc00008e780)({
			ResultsByTime: [{
				Estimated: true,
				Groups: [],
				TimePeriod: {
				  End: "2018-12-25",
				  Start: "2018-12-24"
				},
				Total: {
				  BlendedCost: {
					Amount: "143.3628158149",
					Unit: "USD"
				  }
				}
			  }]
		  })
	*/
	err := ctx.reporter.Report(opt, func(out *costexplorer.GetCostAndUsageOutput) error {
		results := out.ResultsByTime
		if len(results) != 1 {
			return fmt.Errorf("aws cost of yesterday: want just one Output.ResultsByTime, got %d results", len(results))
		}
		ctx.CostOfYesterday = results[0]
		return nil
	})
	if err != nil {
		ctx.Err = fmt.Errorf("cost of yesterday: %s", err)
	}

	return ctx
}

func (ctx *awsDailyReportContext) monthlySumOfCost() *awsDailyReportContext {
	if ctx.Err != nil {
		return ctx
	}
	year, month, _ := time.Now().UTC().Date()
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC) // 0日にすると先月扱いになる
	end := start.AddDate(0, 1, 0).AddDate(0, 0, -1)
	opt := &cost.AWSReportOptions{
		Start:       start,
		End:         end,
		Granularity: cost.Monthly,
	}
	/*
		(*costexplorer.ResultByTime)(0xc00046e150)({
			Estimated: true,
			Groups: [],
			TimePeriod: {
			  End: "2018-12-31",
			  Start: "2018-12-01"
			},
			Total: {
			  BlendedCost: {
				Amount: "3956.3458977203",
				Unit: "USD"
			  }
			}
		  })
	*/
	err := ctx.reporter.Report(opt, func(out *costexplorer.GetCostAndUsageOutput) error {
		results := out.ResultsByTime

		// Estimated => falseな値が返却されるため
		var estimated []*costexplorer.ResultByTime
		for _, result := range results {
			if aws.BoolValue(result.Estimated) {
				estimated = append(estimated, result)
			}
		}
		if len(estimated) != 1 {
			return fmt.Errorf("aws monthly sum of cost: want just one Output.ResultsByTime, got %d results", len(results))
		}
		ctx.MonthlySumOfCost = estimated[0]
		return nil
	})
	if err != nil {
		ctx.Err = fmt.Errorf("failed to monthly sum: %s", err)
	}
	return ctx
}

func (ctx *awsDailyReportContext) monthlyForcast() *awsDailyReportContext {
	if ctx.Err != nil {
		return ctx
	}
	// その月の最後の日の場合はskipする
	// Forcaseの仕様でかならず次の日以降をStartに指定しなければならないが、来月の予想は必要ない

	start := time.Now().UTC().AddDate(0, 0, 1)
	year, month, day := time.Now().UTC().Date()
	beginOfMonth := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC) // 0日にすると先月扱いになる
	endOfMonth := beginOfMonth.AddDate(0, 1, -1)
	opt := &cost.AWSReportOptions{
		Start:       start,
		End:         endOfMonth,
		Granularity: cost.Monthly,
	}

	if day == endOfMonth.Day() {
		fmt.Println("today is endOfMonth. skip monthly forcast")
		ctx.SkipMonthlyForcast = true
		ctx.MonthlyForcast = &costexplorer.GetCostForecastOutput{}
		return ctx
	}

	/*
			(*costexplorer.GetCostForecastOutput)(0xc0000af280)({
		  ForecastResultsByTime: [{
		      MeanValue: "4815.499082616509",
		      PredictionIntervalLowerBound: "4794.605026549281",
		      PredictionIntervalUpperBound: "4836.393138683738",
		      TimePeriod: {
		        End: "2019-01-01",
		        Start: "2018-12-01"
		      }
		    }],
		  Total: {
		    Amount: "4815.499082616509",
		    Unit: "USD"
		  }
		})
	*/
	err := ctx.reporter.Forecast(opt, func(out *costexplorer.GetCostForecastOutput) error {
		ctx.MonthlyForcast = out
		return nil
	})
	if err != nil {
		ctx.Err = fmt.Errorf("failed to monthly forcast: %s", err)
	}
	return ctx
}

func (ctx *awsDailyReportContext) notify() *awsDailyReportContext {
	if ctx.Err != nil {
		ctx.notifyErr()
		return ctx
	}
	ctx.Err = ctx.notifyReport()
	return ctx
}

func (ctx *awsDailyReportContext) notifyErr() {
	payload := ctx.errPayload()
	ctx.doNotify(payload)
}

func (ctx *awsDailyReportContext) notifyReport() error {
	// spew.Dump(ctx.CostOfYesterday)
	// spew.Dump(ctx.MonthlySumOfCost)
	// spew.Dump(ctx.MonthlyForcast)
	payload := ctx.payload()
	return ctx.doNotify(payload)

}

func (ctx *awsDailyReportContext) doNotify(payload *webhook.Payload) error {
	return ctx.notificator.SendPayload(payload)
}

func (ctx *awsDailyReportContext) payload() *webhook.Payload {
	titlize := func(r *costexplorer.ResultByTime) string {
		start := aws.StringValue(r.TimePeriod.Start)
		end := aws.StringValue(r.TimePeriod.End)
		return fmt.Sprintf("%s ~ %s", start, end)
	}
	formatValue := func(r *costexplorer.ResultByTime) string {
		const metric = "BlendedCost" // FIX THIS HARD CODE
		amount := StripDot(aws.StringValue(r.Total[metric].Amount))
		unit := Emojify(aws.StringValue(r.Total[metric].Unit))
		return fmt.Sprintf("%s%s", unit, amount)
	}
	formatForecast := func(mv *costexplorer.MetricValue) string {
		if ctx.SkipMonthlyForcast || mv == nil {
			return "skip forecast at the end of month"
		}
		amount := StripDot(aws.StringValue(mv.Amount))
		unit := Emojify(aws.StringValue(mv.Unit))
		return fmt.Sprintf("%s%s", unit, amount)
	}
	payload := &webhook.Payload{
		Text: ":chart:  AWS Cost Report",
		Attachments: []*webhook.Attachment{
			{
				Color: "good",
				Fields: []*webhook.Field{
					{
						Title: titlize(ctx.CostOfYesterday),
						Value: formatValue(ctx.CostOfYesterday),
						Short: false,
					},
					{
						Title: titlize(ctx.MonthlySumOfCost),
						Value: formatValue(ctx.MonthlySumOfCost),
						Short: false,
					},
					{
						Title: "forecast",
						Value: formatForecast(ctx.MonthlyForcast.Total),
						Short: false,
					},
				},
			},
		},
	}
	return payload
}

func (ctx *awsDailyReportContext) errPayload() *webhook.Payload {
	payload := &webhook.Payload{
		Text: "error occured",
		Attachments: []*webhook.Attachment{
			{
				Color: "danger",
				Text:  ctx.Err.Error(),
			},
		},
	}
	return payload
}

// StripDot strip dot from unit. ex 4075.123456 -> 4075
func StripDot(amount string) string {
	dot := strings.Index(amount, ".")
	if dot == -1 {
		return amount
	}
	return amount[:dot]
}

// Emojify convert USD -> :heavy_dollar_sign:
func Emojify(unit string) string {
	if strings.TrimSpace(unit) == "USD" {
		return "$"
	}
	return unit
}
