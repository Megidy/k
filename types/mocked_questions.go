package types

var MockedQuestions = []Question{
	{
		Id: "q1",
		Topic: &Topic{
			TopicID: "geo1",
			UserID:  "user789",
			Name:    "World Geography",
		},
		Type:          "image",
		ImageLink:     "https://example.com/mountains.jpg",
		Question:      "What is the highest mountain in the world?",
		Answers:       []string{"Mount Everest", "K2", "Kilimanjaro", "Denali"},
		CorrectAnswer: "Mount Everest",
	},
	{
		Id: "q2",
		Topic: &Topic{
			TopicID: "geo2",
			UserID:  "user987",
			Name:    "Countries & Capitals",
		},
		Type:          "image",
		ImageLink:     "https://example.com/capitals.jpg",
		Question:      "What is the capital of Canada?",
		Answers:       []string{"Toronto", "Vancouver", "Ottawa", "Montreal"},
		CorrectAnswer: "Ottawa",
	},
	{
		Id: "q3",
		Topic: &Topic{
			TopicID: "geo2",
			UserID:  "user987",
			Name:    "question3",
		},
		Type:          "image",
		ImageLink:     "https://example.com/capitals.jpg",
		Question:      "what is question 3",
		Answers:       []string{"1", "2", "3", "4"},
		CorrectAnswer: "3",
	},
}
