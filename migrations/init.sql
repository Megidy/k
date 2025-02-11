CREATE DATABASE IF NOT EXISTS k;
use k;

DROP TABLE IF EXISTS users, topics, questions, answers;

CREATE TABLE IF NOT EXISTS users(
	id varchar(255),
    username text,
    email text,
    password varchar(255),
    description varchar(255) NOT NULL default("no description"),
	profile_picture varchar(255),
    PRIMARY KEY(id)
   
);

CREATE TABLE IF NOT EXISTS topics(
	topic_id VARCHAR(255),
    user_id varchar(255),
    name varchar(255),
    PRIMARY KEY(name,user_id),
	UNIQUE (name, user_id)
);

CREATE TABLE IF NOT EXISTS questions(
    id VARCHAR(255),
    topic_id VARCHAR(255),
    type ENUM('text','image') DEFAULT 'text',
    image_link VARCHAR(255) NOT NULL DEFAULT '', 
    question TEXT,
    correct_answer TEXT,
    PRIMARY KEY(id)
);

CREATE TABLE IF NOT EXISTS answers(
	question_id VARCHAR(255),
	answer TEXT
    
);



INSERT INTO topics (topic_id, user_id, name) 
VALUES ('geo_001', 'public', 'Geography');

INSERT INTO questions (id, topic_id, type, image_link, question, correct_answer) VALUES
('q1', 'geo_001', 'text', '', 'What is the capital of France?', 'Paris'),
('q2', 'geo_001', 'text', '', 'Which is the largest ocean in the world?', 'Pacific Ocean'),
('q3', 'geo_001', 'text', '', 'Mount Everest is located in which mountain range?', 'Himalayas'),
('q4', 'geo_001', 'text', '', 'Which continent has the most countries?', 'Africa'),
('q5', 'geo_001', 'text', '', 'What is the longest river in the world?', 'Nile River'),
('q6', 'geo_001', 'text', '', 'Which country has the most natural lakes?', 'Canada'),
('q7', 'geo_001', 'text', '', 'What is the name of the largest desert in the world?', 'Antarctic Desert'),
('q8', 'geo_001', 'text', '', 'Which country is known as the Land of the Rising Sun?', 'Japan'),
('q9', 'geo_001', 'text', '', 'What is the smallest country in the world?', 'Vatican City'),
('q10', 'geo_001', 'text', '', 'Which European country has the most islands?', 'Sweden'),
('q11', 'geo_001', 'text', '', 'What is the capital of Australia?', 'Canberra'),
('q12', 'geo_001', 'text', '', 'Which U.S. state has the longest coastline?', 'Alaska'),
('q13', 'geo_001', 'text', '', 'Which country is both in Europe and Asia?', 'Turkey'),
('q14', 'geo_001', 'text', '', 'What is the name of the largest freshwater lake by volume?', 'Lake Baikal'),
('q15', 'geo_001', 'text', '', 'What is the highest waterfall in the world?', 'Angel Falls'),
('q16', 'geo_001', 'text', '', 'Which continent has the highest population?', 'Asia'),
('q17', 'geo_001', 'text', '', 'What is the deepest ocean in the world?', 'Pacific Ocean'),
('q18', 'geo_001', 'text', '', 'Which country has the most volcanoes?', 'Indonesia'),
('q19', 'geo_001', 'text', '', 'Which African country has the highest population?', 'Nigeria'),
('q20', 'geo_001', 'text', '', 'What is the capital of Brazil?', 'Brasília');


INSERT INTO answers (question_id, answer) VALUES
('q1', 'Paris'),
('q1', 'London'),
('q1', 'Berlin'),
('q1', 'Rome'),
('q1', 'Madrid'),
('q2', 'Pacific Ocean'),
('q2', 'Atlantic Ocean'),
('q2', 'Indian Ocean'),
('q2', 'Arctic Ocean'),
('q2', 'Southern Ocean'),
('q3', 'Himalayas'),
('q3', 'Rockies'),
('q3', 'Andes'),
('q3', 'Alps'),
('q3', 'Appalachians'),
('q4', 'Africa'),
('q4', 'Asia'),
('q4', 'Europe'),
('q4', 'North America'),
('q4', 'Australia'),
('q5', 'Nile River'),
('q5', 'Amazon River'),
('q5', 'Yangtze River'),
('q5', 'Mississippi River'),
('q5', 'Ganges River'),
('q6', 'Canada'),
('q6', 'USA'),
('q6', 'Mexico'),
('q6', 'Brazil'),
('q6', 'Argentina'),
('q7', 'Antarctic Desert'),
('q7', 'Sahara Desert'),
('q7', 'Gobi Desert'),
('q7', 'Karakum Desert'),
('q7', 'Sonoran Desert'),
('q8', 'Japan'),
('q8', 'China'),
('q8', 'South Korea'),
('q8', 'North Korea'),
('q8', 'Taiwan'),
('q9', 'Vatican City'),
('q9', 'Monaco'),
('q9', 'San Marino'),
('q9', 'Liechtenstein'),
('q9', 'Malta'),
('q10', 'Sweden'),
('q10', 'Norway'),
('q10', 'Finland'),
('q10', 'Denmark'),
('q10', 'Iceland'),
('q11', 'Canberra'),
('q11', 'Sydney'),
('q11', 'Melbourne'),
('q11', 'Brisbane'),
('q11', 'Perth'),
('q12', 'Alaska'),
('q12', 'Greenland'),
('q12', 'Siberia'),
('q12', 'Nordic Countries'),
('q12', 'Faroe Islands'),
('q13', 'Turkey'),
('q13', 'Iran'),
('q13', 'Syria'),
('q13', 'Iraq'),
('q13', 'Jordan'),
('q14', 'Lake Baikal'),
('q14', 'Lake Tanganyika'),
('q14', 'Lake Superior'),
('q14', 'Lake Victoria'),
('q14', 'Lake Michigan'),
('q15', 'Angel Falls'),
('q15', 'Victoria Falls'),
('q15', 'Niagara Falls'),
('q15', 'Iguazu Falls'),
('q15', 'Goðafoss'),
('q16', 'Asia'),
('q16', 'Europe'),
('q16', 'North America'),
('q16', 'Africa'),
('q16', 'Australia'),
('q17', 'Pacific Ocean'),
('q17', 'Atlantic Ocean'),
('q17', 'Indian Ocean'),
('q17', 'Arctic Ocean'),
('q17', 'Southern Ocean'),
('q18', 'Indonesia'),
('q18', 'Malaysia'),
('q18', 'Thailand'),
('q18', 'Philippines'),
('q18', 'Vietnam'),
('q19', 'Nigeria'),
('q19', 'Kenya'),
('q19', 'South Africa'),
('q19', 'Egypt'),
('q19', 'Morocco'),
('q20', 'Brasília'),
('q20', 'Rio de Janeiro'),
('q20', 'São Paulo'),
('q20', 'Salvador'),
('q20', 'Fortaleza');
