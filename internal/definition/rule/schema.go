// Copyright Splunk, Inc.
// SPDX-License-Identifier: MPL-2.0

package rule

import (
	"fmt"
	"hash/crc32"
	"io"
	"slices"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/splunk-terraform/terraform-provider-signalfx/internal/check"
)

func NewSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"severity": {
			Type:             schema.TypeString,
			Required:         true,
			ValidateDiagFunc: check.SeverityLevel(),
			Description:      "The severity of the rule, must be one of: Critical, Warning, Major, Minor, Info",
		},
		"detect_label": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "A detect label which matches a detect label within the program text",
		},
		"description": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Description of the rule",
		},
		"notifications": {
			Type:     schema.TypeList,
			Optional: true,
			Elem: &schema.Schema{
				Type:             schema.TypeString,
				ValidateDiagFunc: check.Notification(),
			},
			Description: "List of strings specifying where notifications will be sent when an incident occurs. See https://developers.signalfx.com/v2/docs/detector-model#notifications-models for more info",
		},
		"disabled": {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     false,
			Description: "(default: false) When true, notifications and events will not be generated for the detect label",
		},
		"parameterized_body": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom notification message body when an alert is triggered. See https://developers.signalfx.com/v2/reference#detector-model for more info",
		},
		"parameterized_subject": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Custom notification message subject when an alert is triggered. See https://developers.signalfx.com/v2/reference#detector-model for more info",
		},
		"runbook_url": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "URL of page to consult when an alert is triggered",
		},
		"tip": {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Plain text suggested first course of action, such as a command to execute.",
		},
	}
}

func Hash(v any) int {
	var (
		hash = crc32.NewIEEE()
		rule = v.(map[string]any)
	)

	for _, field := range []string{
		"description",
		"severity",
		"detect_label",
		"disabled",
	} {
		_, _ = io.WriteString(hash, fmt.Sprintf("%s-", rule[field]))
	}

	for _, field := range []string{
		"parameterized_body",
		"parameterized_subject",
		"runbook_url",
		"tip",
	} {
		if s, ok := rule[field]; ok {
			_, _ = io.WriteString(hash, fmt.Sprintf("%s-", s))
		}
	}

	if notifys, ok := rule["notifications"].([]any); ok {
		slices.SortFunc(notifys, func(a, b any) int {
			return strings.Compare(a.(string), b.(string))
		})

		for _, n := range notifys {
			_, _ = io.WriteString(hash, fmt.Sprintf("%s-", n))
		}
	}

	// crc32 returns a uint32, but for our use we need
	// and non negative integer. Here we cast to an integer
	// and invert it if the result is negative.
	code := int(hash.Sum32())
	if code < 0 {
		code = -code
	}
	return code
}
