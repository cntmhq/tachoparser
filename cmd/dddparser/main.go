package main

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"time"

	_ "github.com/kyburz-switzerland-ag/tachoparser/internal/pkg/certificates"
	"github.com/kyburz-switzerland-ag/tachoparser/pkg/decoder"
	"github.com/rs/zerolog"
)

var (
	card   = flag.Bool("card", false, "File is a driver card")
	vu     = flag.Bool("vu", false, "File is a vu file")
	input  = flag.String("input", "", "Input file (optional, stdin is used if not set)")
	output = flag.String("output", "", "Output file (optional, stdout is used if not set)")
)

func MinutesSinceMidnightToTimestamp(minutes int) string {
	// Define the base time as midnight UTC
	baseTime := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	// Add the minutes to the base time
	timestamp := baseTime.Add(time.Duration(minutes) * time.Minute)

	// Format the timestamp as HH:MM
	formattedTime := timestamp.Format("15:04 -0700 MST") // 24-hour clock format
	return formattedTime
}

func main() {
	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Printf("loaded certificates: %v %v", len(decoder.PKsFirstGen), len(decoder.PKsSecondGen))

	flag.Parse()
	if (*card && *vu) || (!*card && !*vu) {
		log.Fatal().Msg("either card or vu must be set")
	}

	var data []byte
	if *input == "" {
		var err error
		data, err = io.ReadAll(os.Stdin)
		if err != nil {
			log.Fatal().Msgf("error: could not read stdin: %v", err)
		}
	} else {
		var err error
		data, err = os.ReadFile(*input)
		if err != nil {
			log.Fatal().Msgf("error: could not read file: %v", err)
		}
	}

	var dataOut []byte
	if *card {
		var err error
		var c decoder.Card
		_, err = decoder.UnmarshalTLV(data, &c)
		if err != nil {
			log.Fatal().Msgf("error: could not parse card: %v", err)
		}
		dataOut, err = json.Marshal(c)
		if err != nil {
			log.Fatal().Msgf("error: could not marshal card: %v", err)
		}
	} else {
		var err error
		var v decoder.Vu
		_, err = decoder.UnmarshalTV(data, &v)
		if err != nil {
			log.Fatal().Msgf("error: could not parse vu data: %v", err)
		}

		for _, vu_activity_gen_2_v2_range := range v.VuActivitiesSecondGenV2 {
			//log.Debug().Interface("vu_activity_gen_2_v_2", vu_activity_gen_2_v2_range).Send()

			log.Debug().Msgf("record_download_date: %v", vu_activity_gen_2_v2_range.DateOfDayDownloadedRecordArray.Records[0].Decode())

			for _, vu_card_iw := range vu_activity_gen_2_v2_range.VuCardIWRecordArray.Records {
				log.Debug().Msgf("card_insertion: %v", vu_card_iw.CardInsertionTime.Decode())
				log.Debug().Msgf("card_withdrawal: %v", vu_card_iw.CardWithdrawalTime.Decode())
			}
			for _, vu_activity := range vu_activity_gen_2_v2_range.VuActivityDailyRecordArray.Records {
				log.Debug().Bool("card_present", vu_activity.Decode().CardPresent).Bool("driver_present", vu_activity.Decode().Driver).Bool("team", vu_activity.Decode().Team).Int("work_type", int(vu_activity.Decode().WorkType)).Str("time", MinutesSinceMidnightToTimestamp(vu_activity.Decode().Minutes)).Send()
			}

		}

		dataOut, err = json.Marshal(v)

		if err != nil {
			log.Fatal().Msgf("error: could not marshal vu data: %v", err)
		}
		// log.Debug().RawJSON("data", dataOut).Send()
	}
	if *output == "" || *output == "-" {
	} else {
		err := os.WriteFile(*output, dataOut, 0644)
		if err != nil {
			log.Fatal().Msgf("error: could not write output file: %v", err)
		}
	}
}
