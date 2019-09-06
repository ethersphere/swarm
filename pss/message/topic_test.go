package message_test

import (
	"fmt"
	"testing"

	"github.com/epiclabs-io/ut"
	"github.com/ethersphere/swarm/pss/message"
)

// test that topic conversion functions give predictable results
var someTopics = []string{"These", "are", "some", "topics", "A topic can be very long as well, longer than TopicLength"}

func TestTopic(tx *testing.T) {
	t := ut.BeginTest(tx, false) // set to true to generate test results
	defer t.FinishTest()

	for i, topicString := range someTopics {
		topic := message.NewTopic([]byte(topicString))

		// Test marshalling and unmarshalling the topic as JSON:
		t.TestJSONMarshaller(fmt.Sprintf("topic%d.json", i), topic)

		// test Stringer:
		s := topic.String()
		t.EqualsKey(fmt.Sprintf("topic%d", i), s)
	}
}
