package evt

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/forstmeier/todalytics/pkg/tbl"
)

func TestNew(t *testing.T) {
	client := New("authorizationToken")
	if client == nil {
		t.Error("error creating evt client")
	}
}

type mockHelp struct {
	mockGetExtraValuesOutput *extraValues
	mockGetExtraValuesError  error
}

func (m *mockHelp) getExtraValues(ctx context.Context, projectID, itemID int, sectionID *int, labelIDs []int) (*extraValues, error) {
	return m.mockGetExtraValuesOutput, m.mockGetExtraValuesError
}

func TestConvert(t *testing.T) {
	timeValue, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Fatalf("error creating test time value: %s", err.Error())
	}

	tests := []struct {
		description              string
		data                     []byte
		mockGetExtraValuesOutput *extraValues
		mockGetExtraValuesError  error
		row                      *tbl.Row
		function                 string
		error                    error
	}{
		{
			description:              "error unmarshalling input",
			data:                     []byte("---------"),
			mockGetExtraValuesOutput: nil,
			mockGetExtraValuesError:  nil,
			function:                 "json unmarshal",
			error:                    &ConvertError{},
		},
		{
			description:              "error get extra values",
			data:                     []byte(`{"event_data": {"section_id": 1}}`),
			mockGetExtraValuesOutput: nil,
			mockGetExtraValuesError:  errors.New("mock get extra values error"),
			row:                      nil,
			function:                 "get extra values",
			error:                    &ConvertError{},
		},
		{
			description:              "error parse date added",
			data:                     []byte(`{"event_data": {"section_id": 1, "date_added": "---"}}`),
			mockGetExtraValuesOutput: &extraValues{},
			mockGetExtraValuesError:  nil,
			row:                      nil,
			function:                 "parse times",
			error:                    &ConvertError{},
		},
		{
			description: "successful convert invocation",
			data:        []byte(`{"event_data": {"section_id": 1, "date_added": "2006-01-02T15:04:05Z", "date_completed": "2006-01-02T15:04:05Z", "labels": [1]}}`),
			mockGetExtraValuesOutput: &extraValues{
				projectName: "project_name",
				notes:       []string{"note content"},
				sectionName: *stringPointer(t, "section_name"),
				labelNames:  []string{"label"},
			},
			mockGetExtraValuesError: nil,
			row: &tbl.Row{
				ItemID:        0,
				Event:         "",
				UserID:        0,
				UserEmail:     "",
				ProjectID:     0,
				ProjectName:   "project_name",
				Content:       "",
				Description:   "",
				Notes:         []string{"note content"},
				Priority:      0,
				ParentID:      nil,
				SectionID:     intPointer(t, 1),
				SectionName:   stringPointer(t, "section_name"),
				LabelIDs:      []int{1},
				LabelNames:    []string{"label"},
				Checked:       false,
				DateAdded:     timeValue,
				DateCompleted: &timeValue,
			},
			function: "",
			error:    nil,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			client := &Client{
				helper: &mockHelp{
					mockGetExtraValuesOutput: test.mockGetExtraValuesOutput,
					mockGetExtraValuesError:  test.mockGetExtraValuesError,
				},
			}

			row, err := client.Convert(context.Background(), test.data)

			if err != nil {
				switch e := test.error.(type) {
				case *ConvertError:
					if !errors.As(err, &e) {
						t.Errorf("incorrect error, received: %v, expected: %v", err, e)
					}
				default:
					t.Fatalf("unexpected error type: %v", err)
				}
			} else {
				test.row.ID = row.ID               // quick solution to pass the deep equal check
				test.row.Timestamp = row.Timestamp // quick solution to pass the deep equal check
				if !reflect.DeepEqual(row, test.row) {
					t.Errorf("incorrect row, \nreceived: %+v,\nexpected: %+v", row, test.row)
				}
			}
		})
	}
}

func stringPointer(t *testing.T, input string) *string {
	return &input
}

func intPointer(t *testing.T, input int) *int {
	return &input
}
