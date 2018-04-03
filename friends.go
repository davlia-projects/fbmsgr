package fbmsgr

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/unixpickle/essentials"
)

type FriendInfo struct {
	ID             json.Number `json:"id"`
	AlternateName  string      `json:"alternateName"`
	FirstName      string      `json:"firstName"`
	IsFriend       string      `json:"isFriend"`
	FullName       string      `json:"fullName"`
	ProfilePicture string      `json:"profilePicture"`
	Type           string      `json:"type"`
	ProfileURL     string      `json:"profileUrl"`
	Vanity         string      `json:"vanity"`
	IsBirthday     string      `json:"isBirthday"`
}

func (s *Session) Friend(fbid string) (res *FriendInfo, err error) {
	defer essentials.AddCtxTo("fbmsgr: friends", &err)
	params, err := s.commonParams()
	params.Set("ids[0]", fbid)
	if err != nil {
		return nil, err
	}
	reqURL := BaseURL + "/chat/user_info"
	body, err := s.jsonForPost(reqURL, params)
	if err != nil {
		return nil, err
	}
	var respObj struct {
		Payload struct {
			Profiles map[string]*FriendInfo `json:"profiles"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		log.Printf("%+v\n", err)
		return nil, errors.New("parse json: " + err.Error())
	}
	return respObj.Payload.Profiles[fbid], nil
}

func (s *Session) Friends() (res map[string]*FriendInfo, err error) {
	defer essentials.AddCtxTo("fbmsgr: friends", &err)
	params, err := s.commonParams()
	params.Set("viewer", s.userID)
	if err != nil {
		return nil, err
	}
	reqURL := BaseURL + "/chat/user_info_all"
	body, err := s.jsonForPost(reqURL, params)
	if err != nil {
		return nil, err
	}
	var respObj struct {
		Payload map[string]*FriendInfo `json:"payload"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		log.Printf("%+v\n", err)
		return nil, errors.New("parse json: " + err.Error())
	}
	return respObj.Payload, nil
}
