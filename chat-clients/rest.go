package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/aasmall/dicemagic/internal/dicelang/errors"
	pb "github.com/aasmall/dicemagic/internal/proto"
)

type RESTRollResponse struct {
	Cmd    string `json:"cmd"`
	Result string `json:"result"`
	Ok     bool   `json:"ok"`
	Err    string `json:"err,omitempty"`
}
type RESTRollRequest struct {
	Cmd         string `json:"cmd"`
	Chart       bool   `json:"with_chart,omitempty"`
	Probability bool   `json:"with_probability,omitempty"`
}

type Dice struct {
	Count                int64             `protobuf:"varint,1,opt,name=Count,proto3" json:"Count,omitempty"`
	Sides                int64             `protobuf:"varint,2,opt,name=Sides,proto3" json:"Sides,omitempty"`
	Total                int64             `protobuf:"varint,3,opt,name=Total,proto3" json:"Total,omitempty"`
	Faces                []int64           `protobuf:"varint,4,rep,packed,name=Faces,proto3" json:"Faces,omitempty"`
	Color                string            `protobuf:"bytes,5,opt,name=Color,proto3" json:"Color,omitempty"`
	Max                  int64             `protobuf:"varint,6,opt,name=Max,proto3" json:"Max,omitempty"`
	Min                  int64             `protobuf:"varint,7,opt,name=Min,proto3" json:"Min,omitempty"`
	DropHighest          int64             `protobuf:"varint,8,opt,name=DropHighest,proto3" json:"DropHighest,omitempty"`
	DropLowest           int64             `protobuf:"varint,9,opt,name=DropLowest,proto3" json:"DropLowest,omitempty"`
	Chart                []byte            `protobuf:"bytes,10,opt,name=Chart,proto3" json:"Chart,omitempty"`
	Probabilities        map[int64]float64 `protobuf:"bytes,11,rep,name=Probabilities,proto3" json:"Probabilities,omitempty" protobuf_key:"varint,1,opt,name=key,proto3" protobuf_val:"fixed64,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func RESTRollHandler(e interface{}, w http.ResponseWriter, r *http.Request) error {
	env, _ := e.(*env)
	log := env.log.WithRequest(r)
	req := &RESTRollRequest{}

	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		log.Errorf("Unexpected error decoding REST request: %+v", err)
		return err
	}
	resp := &RESTRollResponse{Cmd: req.Cmd}
	diceServerResponse, err := Roll(env.diceServerClient, req.Cmd, RollOptionWithProbability(req.Probability), RollOptionWithChart(req.Chart))
	if err != nil {
		errString := fmt.Sprintf("Unexpected error: %+v", err)
		resp.Ok = false
		resp.Err = errString
		env.log.Error(errString)
		return nil
	}
	if diceServerResponse.Ok {
		resp.Ok = true
		resp.Result = StringFromRollResponse(diceServerResponse)
	} else {
		if diceServerResponse.Error.Code == errors.Friendly {
			resp.Ok = true
			resp.Result = diceServerResponse.Error.Msg
		} else {
			resp.Ok = false
			resp.Err = diceServerResponse.Error.Msg
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
	return nil
}

func StringFromRollResponse(rr *pb.RollResponse) string {
	var s []string
	fmt.Printf("\nROLL RESPONSE\n", rr.String())
	for _, ds := range rr.DiceSets {
		fmt.Printf("\nDICE SET\n", ds.String())
		var faces []interface{}
		for _, d := range ds.Dice {
			fmt.Printf("\nDICE\n", d.String())
			faces = append(faces, facesSliceString(d.Faces))
		}
		s = append(s, fmt.Sprintf("%s = *%s*", fmt.Sprintf(ds.ReString, faces...), strconv.FormatInt(ds.Total, 10)))
	}
	if len(rr.DiceSets) > 1 {
		s = append(s, fmt.Sprintf("Total: %s", strconv.FormatInt(rr.DiceSet.Total, 10)))
	}
	return strings.Join(s, "\n")
}
