package fbmsgr

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/unixpickle/essentials"
)

const actionBufferSize = 500
const actionLogDocID = "1547392382048831"
const threadDocID = "1349387578499440"

// ThreadInfo stores information about a chat thread.
// A chat thread is facebook's internal name for a
// conversation (either a group chat or a 1-on-1).

type ThreadResponse struct {
	ThreadKey struct {
		ThreadFbid  string `json:"thread_fbid"`
		OtherUserID string `json:"other_user_id"`
	} `json:"thread_key"`
	Name        string `json:"name"`
	LastMessage struct {
		Nodes []struct {
			Snippet       string `json:"snippet"`
			MessageSender struct {
				MessagingActor struct {
					ID string `json:"id"`
				} `json:"messaging_actor"`
			} `json:"message_sender"`
			TimestampPrecise string `json:"timestamp_precise"`
		} `json:"nodes"`
	} `json:"last_message"`
	UnreadCount        int    `json:"unread_count"`
	MessagesCount      int    `json:"messages_count"`
	UpdatedTimePrecise string `json:"updated_time_precise"`
	IsPinProtected     bool   `json:"is_pin_protected"`
	IsViewerSubscribed bool   `json:"is_viewer_subscribed"`
	ThreadQueueEnabled bool   `json:"thread_queue_enabled"`
	Folder             string `json:"folder"`
	HasViewerArchived  bool   `json:"has_viewer_archived"`
	IsPageFollowUp     bool   `json:"is_page_follow_up"`
	CannotReplyReason  string `json:"cannot_reply_reason"`
	EphemeralTTLMode   int    `json:"ephemeral_ttl_mode"`
	EventReminders     struct {
		Nodes []interface{} `json:"nodes"`
	} `json:"event_reminders"`
	MontageThread struct {
		ID string `json:"id"`
	} `json:"montage_thread"`
	LastReadReceipt struct {
		Nodes []struct {
			TimestampPrecise string `json:"timestamp_precise"`
		} `json:"nodes"`
	} `json:"last_read_receipt"`
	RelatedPageThread          interface{}   `json:"related_page_thread"`
	AssociatedObject           interface{}   `json:"associated_object"`
	PrivacyMode                int           `json:"privacy_mode"`
	CustomizationEnabled       bool          `json:"customization_enabled"`
	ThreadType                 string        `json:"thread_type"`
	ParticipantAddModeAsString interface{}   `json:"participant_add_mode_as_string"`
	ParticipantsEventStatus    []interface{} `json:"participants_event_status"`
	AllParticipants            struct {
		Nodes []struct {
			MessagingActor struct {
				ID          string `json:"id"`
				Typename    string `json:"__typename"`
				Name        string `json:"name"`
				Gender      string `json:"gender"`
				URL         string `json:"url"`
				BigImageSrc struct {
					URI string `json:"uri"`
				} `json:"big_image_src"`
				ShortName                string      `json:"short_name"`
				Username                 string      `json:"username"`
				IsViewerFriend           bool        `json:"is_viewer_friend"`
				IsMessengerUser          bool        `json:"is_messenger_user"`
				IsVerified               bool        `json:"is_verified"`
				IsMessageBlockedByViewer bool        `json:"is_message_blocked_by_viewer"`
				IsViewerCoworker         bool        `json:"is_viewer_coworker"`
				IsEmployee               interface{} `json:"is_employee"`
			} `json:"messaging_actor"`
		} `json:"nodes"`
	} `json:"all_participants"`
	ReadReceipts struct {
		Nodes []struct {
			Watermark string `json:"watermark"`
			Action    string `json:"action"`
			Actor     struct {
				ID string `json:"id"`
			} `json:"actor"`
		} `json:"nodes"`
	} `json:"read_receipts"`
	DeliveryReceipts struct {
		Nodes []struct {
			TimestampPrecise string `json:"timestamp_precise"`
		} `json:"nodes"`
	} `json:"delivery_receipts"`
}

type ThreadInfo struct {
	ThreadID   string `json:"thread_id"`
	ThreadFBID string `json:"thread_fbid"`
	Name       string `json:"name"`

	// OtherUserFBID is nil for group chats.
	OtherUserFBID *string `json:"other_user_fbid"`

	// Participants contains a list of FBIDs.
	Participants []string `json:"participants"`

	// Snippet stores the last message sent in the thread.
	Snippet       string `json:"snippet"`
	SnippetSender string `json:"snippet_sender"`

	UnreadCount  int `json:"unread_count"`
	MessageCount int `json:"message_count"`

	Timestamp       float64 `json:"timestamp"`
	ServerTimestamp float64 `json:"server_timestamp"`
}

func (t *ThreadInfo) canonicalizeFBIDs() {
	t.SnippetSender = stripFBIDPrefix(t.SnippetSender)
	for i, p := range t.Participants {
		t.Participants[i] = stripFBIDPrefix(p)
	}
}

// ParticipantInfo stores information about a user.
type ParticipantInfo struct {
	// ID is typically "fbid:..."
	ID string `json:"id"`

	FBID   string `json:"fbid"`
	Gender int    `json:"gender"`
	HREF   string `json:"href"`

	ImageSrc    string `json:"image_src"`
	BigImageSrc string `json:"big_image_src"`

	Name      string `json:"name"`
	ShortName string `json:"short_name"`
}

// A ThreadListResult stores the result of listing the
// user's chat threads.
type ThreadListResult struct {
	Threads      []*ThreadInfo      `json:"threads"`
	Participants []*ParticipantInfo `json:"participants"`
}

func (s *Session) Thread(threadID string) (res *ThreadInfo, err error) {
	defer essentials.AddCtxTo("fbmsgr: thread", &err)

	params := map[string]interface{}{
		"id":                 threadID,
		"before":             nil,
		"message_limit":      0,
		"load_messages":      0,
		"load_read_receipts": false,
	}

	var respObj struct {
		MessageThread *ThreadResponse `json:"message_thread"`
	}

	s.graphQLDoc("1498317363570230", params, &respObj)
	thread := marshalThreadInfo(respObj.MessageThread)
	return thread, nil
}

func (s *Session) Threads(limit int) (res *ThreadListResult, err error) {
	defer essentials.AddCtxTo("fbmsgr: threads", &err)

	params := map[string]interface{}{
		"limit":  limit,
		"before": nil,
		"tags":   []string{"INBOX"},
		"includeDeliveryReceipts": true,
		"includeSeqID":            false,
	}

	var respObj struct {
		Viewer struct {
			MessageThreads struct {
				Threads []*ThreadResponse `json:"nodes"`
			} `json:"message_threads"`
		} `json:"viewer"`
	}

	s.graphQLDoc(threadDocID, params, &respObj)
	threads := marshalThreadList(respObj.Viewer.MessageThreads.Threads)

	result := &ThreadListResult{
		Threads:      threads,
		Participants: []*ParticipantInfo{},
	}
	// TODO: handle participants json
	// log.Printf("%s\n", respObj)

	return result, nil
}

func marshalThreadList(respObj []*ThreadResponse) []*ThreadInfo {
	out := []*ThreadInfo{}
	for _, resp := range respObj {
		out = append(out, marshalThreadInfo(resp))
	}
	return out
}

func marshalThreadInfo(resp *ThreadResponse) *ThreadInfo {
	participantIDs := []string{}
	for _, participant := range resp.AllParticipants.Nodes {
		participantIDs = append(participantIDs, participant.MessagingActor.ID)
	}

	threadInfo := &ThreadInfo{
		ThreadID:      canonicalFBID(resp.ThreadKey.ThreadFbid),
		ThreadFBID:    canonicalFBID(resp.ThreadKey.ThreadFbid),
		OtherUserFBID: &resp.ThreadKey.OtherUserID,
		Participants:  participantIDs,
		UnreadCount:   resp.UnreadCount,
		MessageCount:  resp.MessagesCount,
	}
	if resp.Name == "" && len(participantIDs) == 2 {
		threadInfo.Name = resp.AllParticipants.Nodes[0].MessagingActor.Name
	} else if resp.Name == "" {
		threadInfo.Name = "Unnamed Group DM"
	} else {
		threadInfo.Name = resp.Name
	}
	if len(resp.LastMessage.Nodes) > 0 {
		node := resp.LastMessage.Nodes[0]
		threadInfo.Snippet = node.Snippet
		threadInfo.SnippetSender = node.MessageSender.MessagingActor.ID
		timestamp, err := strconv.ParseFloat(node.TimestampPrecise, 64)
		if err != nil {
			timestamp = 0
		}
		threadInfo.Timestamp = timestamp
		threadInfo.ServerTimestamp = timestamp
	}
	return threadInfo
}

// Threads reads a range of the user's chat threads.
// The offset specifiecs the index of the first thread
// to fetch, starting at 0.
// The limit specifies the maximum number of threads.
func (s *Session) ThreadsDeprecated(offset, limit int) (res *ThreadListResult, err error) {
	defer essentials.AddCtxTo("fbmsgr: threads", &err)
	params, err := s.commonParams()
	if err != nil {
		return nil, err
	}
	params.Set("inbox[filter]", "")
	params.Set("inbox[offset]", strconv.Itoa(offset))
	params.Set("inbox[limit]", strconv.Itoa(limit))
	reqURL := BaseURL + "/ajax/mercury/threadlist_info.php?dpr=1"
	body, err := s.jsonForPost(reqURL, params)
	if err != nil {
		return nil, err
	}

	var respObj struct {
		Payload ThreadListResult `json:"payload"`
	}
	if err := json.Unmarshal(body, &respObj); err != nil {
		return nil, errors.New("parse json: " + err.Error())
	}
	for _, x := range respObj.Payload.Participants {
		x.FBID = stripFBIDPrefix(x.FBID)
	}
	for _, x := range respObj.Payload.Threads {
		x.canonicalizeFBIDs()
	}

	return &respObj.Payload, nil
}

// ActionLog reads the contents of a thread.
//
// The fbid parameter specifies the other user ID or the
// group thread ID.
//
// The timestamp parameter specifies the timestamp of the
// earliest action seen from the last call to ActionLog.
// It may be the 0 time, in which case the most recent
// actions will be fetched.
//
// The limit parameter indicates the maximum number of
// actions to fetch.
func (s *Session) ActionLog(fbid string, timestamp time.Time,
	limit int) (log []Action, err error) {
	defer essentials.AddCtxTo("fbmsgr: action log", &err)

	var response struct {
		Thread struct {
			Messages struct {
				Nodes []map[string]interface{} `json:"nodes"`
			} `json:"messages"`
		} `json:"message_thread"`
	}
	params := map[string]interface{}{
		"id":                 fbid,
		"message_limit":      limit,
		"load_messages":      1,
		"load_read_receipts": true,
	}
	if timestamp.IsZero() {
		params["before"] = nil
	} else {
		params["before"] = strconv.FormatInt(timestamp.UnixNano()/1e6, 10)
	}
	if err := s.graphQLDoc(actionLogDocID, params, &response); err != nil {
		return nil, err
	}
	for _, x := range response.Thread.Messages.Nodes {
		log = append(log, decodeAction(x))
	}
	return
}

// FullActionLog fetches all of the actions in a thread
// and returns them in reverse chronological order over
// a channel.
//
// The cancel channel, if non-nil, can be closed to stop
// the fetch early.
//
// The returned channels will both be closed once the
// fetch has completed or been cancelled.
// If an error is encountered during the fetch, it is sent
// over the (buffered) error channel and the fetch will be
// aborted.
func (s *Session) FullActionLog(fbid string, cancel <-chan struct{}) (<-chan Action, <-chan error) {
	if cancel == nil {
		cancel = make(chan struct{})
	}

	res := make(chan Action, actionBufferSize)
	errRes := make(chan error, 1)
	go func() {
		defer close(res)
		defer close(errRes)
		var lastTime time.Time
		var offset int
		for {
			listing, err := s.ActionLog(fbid, lastTime, actionBufferSize)
			if err != nil {
				errRes <- err
				return
			}

			// Remove the one overlapping action.
			if offset > 0 && len(listing) > 0 {
				listing = listing[:len(listing)-1]
			}

			if len(listing) == 0 {
				return
			}

			for i := len(listing) - 1; i >= 0; i-- {
				x := listing[i]
				select {
				case <-cancel:
					return
				default:
				}

				select {
				case res <- x:
				case <-cancel:
					return
				}
			}

			offset += len(listing)
			lastTime = listing[0].ActionTime()
		}
	}()

	return res, errRes
}

// DeleteMessage deletes a message given its ID.
func (s *Session) DeleteMessage(id string) (err error) {
	defer essentials.AddCtxTo("fbmsgr: delete message", &err)

	url := BaseURL + "/ajax/mercury/delete_messages.php?dpr=1"
	values, err := s.commonParams()
	if err != nil {
		return err
	}
	values.Set("message_ids[0]", id)
	_, err = s.jsonForPost(url, values)
	return err
}
