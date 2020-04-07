package message_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/ethersphere/swarm/pss/message"
)

// test that topic conversion functions give predictable results
var someTopics = []string{"These", "are", "some", "topics", "A topic can be very long as well, longer than TopicLength"}
var fixtureTopicHash = []string{`"0x91273d49"`, `"0xba78973d"`, `"0xa6b46dd0"`, `"0xf013aa4b"`, `"0x26f57386"`}
var fixtureTopicStringer = []string{"0x91273d49", "0xba78973d", "0xa6b46dd0", "0xf013aa4b", "0x26f57386"}

func TestTopic(t *testing.T) {

	for i, topicString := range someTopics {
		topic := message.NewTopic([]byte(topicString))

		// Test marshalling and unmarshalling the topic as JSON:
		jsonBytes, err := json.Marshal(topic)
		if err != nil {
			t.Fatal(err)
		}
		expected := fixtureTopicHash[i]
		actual := string(jsonBytes)
		if expected != actual {
			t.Fatalf("Expected JSON serialization to return %s, got %s", expected, actual)
		}

		var topic2 message.Topic
		err = json.Unmarshal(jsonBytes, &topic2)
		if !reflect.DeepEqual(topic, topic2) {
			t.Fatalf("Expected JSON decoding to return the same object, got %v", topic2)
		}

		// test Stringer:
		expected = fixtureTopicStringer[i]
		actual = topic.String()
		if expected != actual {
			t.Fatalf("Expected topic stringer to return %s, got %s", expected, actual)
		}
	}
}
