package main

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"math/rand"
	"net/http"

	"github.com/aasmall/dicemagic/dicelang"

	"github.com/aasmall/dicemagic/roll"
	"google.golang.org/appengine"
)

//SlackRollJSONResponse is the response format for slack slash commands
type SlackRollJSONResponse struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}
type Attachment struct {
	Pretext    string  `json:"pretext"`
	Fallback   string  `json:"fallback"`
	Color      string  `json:"color"`
	AuthorName string  `json:"author_name"`
	Fields     []Field `json:"fields"`
}
type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

func SlackRollHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	ctx := appengine.NewContext(r)
	//Form into syntacticly correct roll statement.
	if r.FormValue("token") != roll.SlackVerificationToken(ctx) {
		fmt.Fprintf(w, "This is not the droid you're looking for.")
		return
	}
	content := fmt.Sprintf("roll %s", r.FormValue("text"))
	stmts, err := dicelang.NewParser(content).Statements()
	if err != nil {
		printErrorToSlack(ctx, err, w, r)
		return
	}
	totalsMap, dice, err := dicelang.GetDiceSets(stmts...)
	if err != nil {
		printErrorToSlack(ctx, err, w, r)
		return
	}
	slackRollResponse := SlackRollJSONResponse{}
	attachment := Attachment{
		Fields: []Field{
			{Title: "foo", Value: "bar"}}}
	if err != nil {
		printErrorToSlack(ctx, err, w, r)
		return
	}
	slackRollResponse.Attachments = append(slackRollResponse.Attachments, attachment)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slackRollResponse)
}

func rollDecisionToSlackAttachment(decision *roll.RollDecision) (Attachment, error) {
	attachment := Attachment{
		Fallback: fmt.Sprintf("I rolled 1d%d to help decide. Results are in: %s",
			len(decision.Choices),
			decision.Choices[decision.Result]),
		AuthorName: decision.Question,
		Color:      stringToColor(decision.Choices[decision.Result])}
	field := Field{Title: decision.Choices[decision.Result], Short: true}
	attachment.Fields = append(attachment.Fields, field)
	return attachment, nil
}

func createSlackAttachment(totalsMap map[string]float64, dice []dicelang.Dice) (Attachment, error) {

}

func rollExpressionToSlackAttachment(expression *roll.RollExpression) (Attachment, error) {
	rollTotals := expression.RollTotals
	attachment := Attachment{
		Fallback:   expression.TotalsString(),
		AuthorName: expression.FormattedString(),
		Color:      stringToColor(expression.InitialText)}

	totalRoll := int64(0)
	allUnspecified := true
	rollCount := 0
	for _, e := range rollTotals {
		if e.RollType != "" {
			allUnspecified = false
		}
		rollCount++
	}
	field := Field{}
	if allUnspecified {
		for _, e := range rollTotals {
			totalRoll += e.RollResult
		}
		field = Field{Title: fmt.Sprintf("%d", totalRoll), Short: false}
		attachment.Fields = append(attachment.Fields, field)
	} else {
		for _, total := range rollTotals {
			totalRoll += total.RollResult
			var fieldTitle string
			if total.RollType == "" {
				fieldTitle = "_Unspecified_"
			} else {
				fieldTitle = total.RollType
			}
			field := Field{
				Title: fmt.Sprintf("%s: %d", fieldTitle, total.RollResult),
				Value: fmt.Sprintf("Rolls: %s", roll.Int64SliceToCSV(total.Faces...)),
				Short: true}
			attachment.Fields = append(attachment.Fields, field)
		}
		if rollCount > 1 {
			field = Field{Title: fmt.Sprintf("For a total of: %d", totalRoll), Short: false}
			attachment.Fields = append(attachment.Fields, field)
		}
	}

	return attachment, nil
}
func createAttachmentsDamageString(rollTotals []roll.Total) string {
	var buffer bytes.Buffer
	for _, e := range rollTotals {
		if e.RollType == "" {
			buffer.WriteString(fmt.Sprintf("%d ", e.RollResult))
		} else {
			buffer.WriteString(fmt.Sprintf("%d of type %s ", e.RollResult, e.RollType))
		}
	}
	return buffer.String()
}
func printErrorToSlack(ctx context.Context, err error, w http.ResponseWriter, r *http.Request) {
	slackRollResponse := new(SlackRollJSONResponse)
	slackRollResponse.Text = err.Error()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slackRollResponse)
}
func stringToColor(input string) string {
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(input))
	hexb := h.Sum(nil)
	hexstr := hex.EncodeToString(hexb[:len(hexb)/2])
	bi.SetString(hexstr, 16)
	rand.Seed(bi.Int64())
	r := rand.Intn(0xff)
	g := rand.Intn(0xff)
	b := rand.Intn(0xff)
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}
