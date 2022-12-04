package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strconv"
	"strings"
	"time"
)

// TODO: 加上内存分配情况分析，来跑 demo
func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	{
		// Contextual Logging
		fmt.Printf("[Contextual Logging]============================================\n")
		// Send(), Msg() 是触发日志输出的关键函数
		log.Debug().
			Str("Scale", "833 cents").
			Float64("Interval", 833.09).
			Msg("Fibonacci is everywhere")

		log.Debug().
			Str("Name", "Tom").
			Send()

		// NOTE: Using `Msgf` generates one allocation even when the logger is disabled.
		//err := errors.New("A repo man spends his life getting into tense situations")
		//service := "myservice"
		//log.Fatal().
		//	Err(err).
		//	Str("service", service).
		//	Msgf("Cannot start %s", service)
	}

	{
		// Leveled Logging
		fmt.Printf("\n[Leveled Logging]============================================\n")
		log.Info().Msg("hello world")       // chain API calling
		log.Log().Str("foo", "bar").Msg("") // log without level
	}

	{
		// Error Logging
		fmt.Printf("\n[Error Logging]============================================\n")
		err := errors.New("seems we have an error here")
		log.Error().Err(err).Msg("")

		// sub error
		// 多个 error 嵌套其实是要 error 本身支持的, Q&A: 看看这个 error 是怎么用的
		//zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		//log.Error().Stack().Err(err).Msg("")
	}

	{
		// Create logger instance to manage different outputs
		fmt.Printf("\n[Create logger instance]============================================\n")
		logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
		logger.Info().Str("foo", "bar").Msg("hello world")

	}

	{
		// Sub-loggers let you chain loggers with additional context
		fmt.Printf("\n[Sub-loggers]============================================\n")
		sublogger := log.With().Str("component", "foo").Logger()
		sublogger.Info().Msg("hello world") // also log 'component:foo'
	}

	{
		// customize format
		fmt.Printf("\n[customize format]============================================\n")
		output := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		output.FormatLevel = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
		}
		output.FormatMessage = func(i interface{}) string {
			return fmt.Sprintf("***%s****", i)
		}
		output.FormatFieldName = func(i interface{}) string {
			return fmt.Sprintf("%s:", i)
		}
		output.FormatFieldValue = func(i interface{}) string {
			return strings.ToUpper(fmt.Sprintf("%s", i))
		}

		customizeLog := zerolog.New(output).With().Timestamp().Logger()

		customizeLog.Info().Str("foo", "bar").Msg("Hello World")
	}

	{
		// Sub dictionary
		fmt.Printf("\n[Sub dictionary]============================================\n")
		log.Info().
			Str("foo", "bar").
			Dict("dict", zerolog.Dict().
				Str("bar", "baz").
				Int("n", 1),
			).Msg("hello world")
		// {"level":"info","foo":"bar","dict":{"bar":"baz","n":1},"time":1670132558,"message":"hello world"}
	}

	{
		// file and line number
		fmt.Printf("\n[file and line number]============================================\n")
		LongLineLog := log.With().Caller().Logger()
		LongLineLog.Info().Msg("hello world")

		// short line logger
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			short := file
			for i := len(file) - 1; i > 0; i-- {
				if file[i] == '/' {
					short = file[i+1:]
					break
				}
			}
			file = short
			return file + ":" + strconv.Itoa(line)
		}
		shortLineLog := log.With().Caller().Logger()
		shortLineLog.Info().Msg("hello world")
	}

	{
		// Log Sampling
		fmt.Printf("\n[Log Sampling]============================================\n")
		sample := uint32(10)
		sampled := log.Sample(&zerolog.BasicSampler{N: sample})
		for i := uint32(0); i < sample+1; i++ {
			sampled.Info().Msg("will be logged every 10 messages")
		}

		/* advanced Sampling */
		// Will let 5 debug messages per period of 1 second.
		// Over 5 debug message, 1 every 100 debug messages are logged.
		// Other levels are not sampled.
		//advanceSampled := log.Sample(zerolog.LevelSampler{
		//	DebugSampler: &zerolog.BurstSampler{
		//		Burst:       5,
		//		Period:      1 * time.Second,
		//		NextSampler: &zerolog.BasicSampler{N: 100},
		//	},
		//})
		//advanceSampled.Debug().Msg("hello world")
	}

	{
		// hooks
		fmt.Printf("\n[hooks]============================================\n")
		hooked := log.Hook(SeverityHook{})
		hooked.Warn().Msg("")
	}

	{
		// Pass a sub-logger by context
		fmt.Printf("\n[Pass a sub-logger by context]============================================\n")
		ctx := log.With().Str("component", "module").Logger().WithContext(context.Background())
		log.Ctx(ctx).Info().Msg("hello world")
	}

	{
		// TODO: interaction with net/http
	}

	{
		// Multiple Log Output
		fmt.Printf("\n[Multiple Log Output]============================================\n")
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		multi := zerolog.MultiLevelWriter(consoleWriter, os.Stdout)
		logger := zerolog.New(multi).With().Timestamp().Logger()
		logger.Info().Msg("Hello World!")
	}
}

type SeverityHook struct{}

func (h SeverityHook) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	if level != zerolog.NoLevel {
		e.Str("severity", level.String())
	}
}
