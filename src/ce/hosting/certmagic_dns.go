package hosting

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/libdns/libdns"
	"github.com/stormkit-io/stormkit-io/src/lib/integrations"
	"github.com/stormkit-io/stormkit-io/src/lib/slog"
	"go.uber.org/zap"
)

type DNSProvider struct {
	awscli *integrations.AWSClient
	zoneID string
}

// NewDNSProviders returns a new instance of DNSProvider.
// This is only used on Stormkit Cloud.
func NewDNSProvider() *DNSProvider {
	awscli, err := integrations.AWS(integrations.ClientArgs{}, nil)

	if err != nil {
		panic(fmt.Sprintf("cannot create aws cli: %s", err.Error()))
	}

	return &DNSProvider{
		awscli: awscli,
		zoneID: os.Getenv("STORMKT_DEV_ZONE_ID"),
	}
}

func (d *DNSProvider) prepareInput(actionType, zone string, record libdns.Record) *route53.ChangeResourceRecordSetsInput {
	slog.Infof("preparing input for obtaining certificates: actionType=%s, zone=%s, record_name=%s", actionType, zone, record.Name)

	return &route53.ChangeResourceRecordSetsInput{
		ChangeBatch: &types.ChangeBatch{
			Changes: []types.Change{
				{
					Action: types.ChangeAction(actionType),
					ResourceRecordSet: &types.ResourceRecordSet{
						Name: aws.String(libdns.AbsoluteName(record.Name, zone)),
						ResourceRecords: []types.ResourceRecord{
							{
								Value: aws.String(record.Value),
							},
						},
						TTL:  aws.Int64(int64(record.TTL)),
						Type: types.RRType(record.Type),
					},
				},
			},
		},
		HostedZoneId: aws.String(d.zoneID),
	}
}

func (d *DNSProvider) AppendRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	slog.Debug(slog.LogOpts{
		Msg:   "appending record for zone id",
		Level: slog.DL1,
		Payload: []zap.Field{
			zap.String("zone", zone),
		},
	})

	var createdRecords []libdns.Record

	for _, record := range records {
		if record.Type == "TXT" {
			record.Value = strconv.Quote(record.Value)
		}

		input := d.prepareInput("UPSERT", zone, record)
		record.TTL = 60

		if _, err := d.awscli.Route53().ChangeResourceRecordSets(ctx, input); err != nil {
			slog.Errorf("error while changing record set=%s, zone=%s, record=%v", err.Error(), zone, record)
			return nil, err
		}

		record.TTL = time.Duration(record.TTL) * time.Second
		createdRecords = append(createdRecords, record)

		slog.Debug(slog.LogOpts{
			Msg:   "created dns record",
			Level: slog.DL1,
			Payload: []zap.Field{
				zap.String("name", record.Name),
				zap.String("id", record.ID),
			},
		})
	}

	return createdRecords, nil
}

// DeleteRecords deletes the records from the zone. If a record does not have an ID,
// it will be looked up. It returns the records that were deleted.
func (d *DNSProvider) DeleteRecords(ctx context.Context, zone string, records []libdns.Record) ([]libdns.Record, error) {
	var deletedRecords []libdns.Record

	slog.Infof("dns-resolver -- delete operation zone=%s, records=%v", zone, records)

	for _, record := range records {
		input := d.prepareInput("DELETE", zone, record)

		if _, err := d.awscli.Route53().ChangeResourceRecordSets(ctx, input); err != nil {
			slog.Errorf("dns-resolver -- failed deleting %v", err)
			return nil, err
		}

		record.TTL = time.Duration(record.TTL) * time.Second
		deletedRecords = append(deletedRecords, record)
	}

	return deletedRecords, nil
}
