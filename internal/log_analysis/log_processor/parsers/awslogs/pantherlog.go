package awslogs

/**
 * Panther is a Cloud-Native SIEM for the Modern Security Team.
 * Copyright (C) 2020 Panther Labs Inc
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

import (
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws/arn"

	"github.com/panther-labs/panther/internal/log_analysis/log_processor/pantherlog"
	"github.com/panther-labs/panther/internal/log_analysis/log_processor/parsers"
)

var (
	awsAccountIDRegex = regexp.MustCompile(`^\d{12}$`)
)

// nolint(lll)
type AWSPantherLog struct {
	parsers.PantherLog

	PantherAnyAWSAccountIds  *parsers.PantherAnyString `json:"p_any_aws_account_ids,omitempty" description:"Panther added field with collection of aws account ids associated with the row"`
	PantherAnyAWSInstanceIds *parsers.PantherAnyString `json:"p_any_aws_instance_ids,omitempty" description:"Panther added field with collection of aws instance ids associated with the row"`
	PantherAnyAWSARNs        *parsers.PantherAnyString `json:"p_any_aws_arns,omitempty" description:"Panther added field with collection of aws arns associated with the row"`
	PantherAnyAWSTags        *parsers.PantherAnyString `json:"p_any_aws_tags,omitempty" description:"Panther added field with collection of aws tags associated with the row"`
}

func (pl *AWSPantherLog) AppendAnyAWSAccountIdPtrs(values ...*string) { // nolint
	for _, value := range values {
		if value != nil {
			pl.AppendAnyAWSAccountIds(*value)
		}
	}
}

func (pl *AWSPantherLog) AppendAnyAWSAccountIds(values ...string) {
	for _, value := range values {
		if !awsAccountIDRegex.MatchString(value) {
			continue
		}
		if pl.PantherAnyAWSAccountIds == nil { // lazy create
			pl.PantherAnyAWSAccountIds = parsers.NewPantherAnyString()
		}
		parsers.AppendAnyString(pl.PantherAnyAWSAccountIds, value)
	}
}

func (pl *AWSPantherLog) AppendAnyAWSInstanceIdPtrs(values ...*string) { // nolint
	for _, value := range values {
		if value != nil {
			pl.AppendAnyAWSInstanceIds(*value)
		}
	}
}

func (pl *AWSPantherLog) AppendAnyAWSInstanceIds(values ...string) {
	if pl.PantherAnyAWSInstanceIds == nil { // lazy create
		pl.PantherAnyAWSInstanceIds = parsers.NewPantherAnyString()
	}
	parsers.AppendAnyString(pl.PantherAnyAWSInstanceIds, values...)
}

func (pl *AWSPantherLog) AppendAnyAWSARNPtrs(values ...*string) {
	for _, value := range values {
		if value != nil {
			pl.AppendAnyAWSARNs(*value)
		}
	}
}

func (pl *AWSPantherLog) AppendAnyAWSARNs(values ...string) {
	if pl.PantherAnyAWSARNs == nil { // lazy create
		pl.PantherAnyAWSARNs = parsers.NewPantherAnyString()
	}
	parsers.AppendAnyString(pl.PantherAnyAWSARNs, values...)
}

func (pl *AWSPantherLog) AppendAnyAWSTagPtrs(values ...*string) {
	for _, value := range values {
		if value != nil {
			pl.AppendAnyAWSTags(*value)
		}
	}
}

// NOTE: value should be of the form <key>:<value>
func (pl *AWSPantherLog) AppendAnyAWSTags(values ...string) {
	if pl.PantherAnyAWSTags == nil { // lazy create
		pl.PantherAnyAWSTags = parsers.NewPantherAnyString()
	}
	parsers.AppendAnyString(pl.PantherAnyAWSTags, values...)
}

const (
	FieldAccountID pantherlog.FieldID = 1000 + iota
	FieldInstanceID
	FieldARN
	FieldTag
)

func init() {
	pantherlog.MustRegisterIndicator(FieldAccountID, pantherlog.FieldMeta{
		NameJSON:    `p_any_aws_account_ids`,
		Name:        `PantherAnyAWSAccountIds`,
		Description: "Panther added field with collection of AWS account ids associated with the row",
	})
	pantherlog.MustRegisterIndicator(FieldInstanceID, pantherlog.FieldMeta{
		NameJSON:    `p_any_aws_instance_ids`,
		Name:        `PantherAnyAWSInstanceIds`,
		Description: "Panther added field with collection of AWS instance ids associated with the row",
	})
	pantherlog.MustRegisterIndicator(FieldARN, pantherlog.FieldMeta{
		NameJSON:    `p_any_aws_arns`,
		Name:        `PantherAnyAWSARNs`,
		Description: "Panther added field with collection of AWS ARNs associated with the row",
	})
	pantherlog.MustRegisterIndicator(FieldTag, pantherlog.FieldMeta{
		NameJSON:    `p_any_aws_tags`,
		Name:        `PantherAnyAWSTags`,
		Description: "Panther added field with collection of AWS tags associated with the row",
	})
	pantherlog.MustRegisterScanner(`aws_arn`, pantherlog.ValueScannerFunc(ScanARN), FieldARN, FieldAccountID, FieldInstanceID)
	pantherlog.MustRegisterScanner(`aws_instance_id`, pantherlog.ValueScannerFunc(ScanInstanceID), FieldInstanceID)
	pantherlog.MustRegisterScanner(`aws_tag`, FieldTag, FieldTag)
	pantherlog.MustRegisterScanner(`aws_account_id`, pantherlog.ValueScannerFunc(ScanAccountID), FieldAccountID)
}

func ScanARN(w pantherlog.ValueWriter, input string) {
	parsedARN, err := arn.Parse(input)
	if err != nil {
		return
	}
	w.WriteValues(FieldARN, input)
	ScanAccountID(w, parsedARN.AccountID)
	scanResourceInstanceID(w, parsedARN.Resource)
}

func scanResourceInstanceID(w pantherlog.ValueWriter, input string) {
	// instanceId: https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/iam-policy-structure.html#EC2_ARN_Format
	if !strings.HasPrefix(input, "instance/") {
		return
	}
	slashIndex := strings.LastIndex(input, "/")
	if slashIndex < len(input)-2 { // not if ends in "/"
		ScanInstanceID(w, input[slashIndex+1:])
	}
}

func ScanAccountID(w pantherlog.ValueWriter, input string) {
	const sizeAccountID = 12
	if len(input) == sizeAccountID && awsAccountIDRegex.MatchString(input) {
		w.WriteValues(FieldAccountID, input)
	}
}

func ScanInstanceID(w pantherlog.ValueWriter, input string) {
	if strings.HasPrefix(input, "i-") {
		w.WriteValues(FieldInstanceID, input)
	}
}
