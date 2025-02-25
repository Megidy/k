package manager

import (
	"log"
	"testing"
	"time"

	"github.com/Megidy/k/config"
	"github.com/Megidy/k/types"
	"github.com/Megidy/k/worker"
)

func TestGameTimeHandler(t *testing.T) {

	workerPool := worker.NewMockedWorkerPool()
	t.Run("should pass if time for answer is up and got restarted", func(t *testing.T) {
		cfg := config.NewMockConfig(2, 2)
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 1, 2, types.MockedQuestions)
		currentTimeLeft := mockedManager.currTime
		expectedResult := cfg.TimeForAnswer - 1
		go mockedManager.GameTimeHandler()

		<-time.After(time.Millisecond * 1010)
		//check if time updates right
		if currentTimeLeft == mockedManager.currTime {
			t.Fatal("expected time to be different ")
		}
		if currentTimeLeft == expectedResult {
			t.Log("current time :", mockedManager.currTime)
			t.Log("number of current questio : ", mockedManager.numberOfCurrentQuestion)
			t.Log("game state : ", mockedManager.gameState)
			t.Fatal("expected time to be different after ~1sec")
		}
		mockedManager.ctx.Done()
	})
	// change of the question triggers restart of the time for answer
	t.Run("Should pass if question changed manually", func(t *testing.T) {
		cfg := config.NewMockConfig(2, 2)
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 1, 2, types.MockedQuestions)
		previousCurrentTime := mockedManager.currTime
		previousCurrentQuestion := mockedManager.numberOfCurrentQuestion
		go mockedManager.Writer()
		go mockedManager.QuestionHandler()
		go mockedManager.GameTimeHandler()

		<-time.After(time.Millisecond * 1010)
		CheckCorrectnessOfTime(t, previousCurrentTime, mockedManager.currTime)
		previousCurrentTime = mockedManager.currTime
		mockedManager.updateQuestionCh <- true
		<-time.After(time.Millisecond * 30)
		CheckCorrectnessOfTime(t, previousCurrentTime, mockedManager.currTime)
		if previousCurrentQuestion == mockedManager.numberOfCurrentQuestion {
			t.Fatal("expected question number be different")
		}
		mockedManager.ctx.Done()
	})

	t.Run("should pass if restartTimeCh called and time was restarted", func(t *testing.T) {
		cfg := config.NewMockConfig(2, 3)
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 1, 2, types.MockedQuestions)
		expectedTime := cfg.TimeForAnswer
		previousCurrentTime := mockedManager.currTime
		//created writer to receive signal from channel writeTimeCh
		go mockedManager.Writer()

		go mockedManager.GameTimeHandler()

		<-time.After(time.Millisecond * 1010)
		CheckCorrectnessOfTime(t, previousCurrentTime, mockedManager.currTime)
		log.Println("current time : ", mockedManager.currTime)
		mockedManager.restartTimeCh <- true
		// it is made to make sure that everything will update.
		<-time.After(time.Millisecond * 10)
		if mockedManager.currTime != expectedTime {
			t.Fatal("expected time to be different after restart, current time: ", mockedManager.currTime, ", expected time:", expectedTime)
		}
		t.Log("passed")
		mockedManager.ctx.Done()
	})
}

func TestBeforeGameLiquidationHandler(t *testing.T) {
	cfg := config.NewMockConfig(1, 1)
	workerPool := worker.NewMockedWorkerPool()
	t.Run("should pass if noone connects to the room", func(t *testing.T) {
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 3, 20, types.MockedQuestions)

		doneCh := make(chan struct{})
		go func() {
			mockedManager.BeforeGameLiquidationHandler()
			close(doneCh)
		}()
		select {
		case <-doneCh:
			//time given for connection to the game, if it ups, than context cancels and game ends
		case <-time.After(time.Millisecond * 1100):
			t.Fatalf("expected context to be canceled")
		}
		err := mockedManager.ctx.Err()
		if err == nil {
			t.Fatalf("expected context to be canceled")
		}
		mockedManager.ctx.Done()
	})
	t.Run("should pass if someone connects to the room", func(t *testing.T) {
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 3, 20, types.MockedQuestions)

		go mockedManager.BeforeGameLiquidationHandler()

		go func() {
			//time for restarting the before game liquidation timer
			time.Sleep(time.Millisecond * 500)
			mockedManager.updateBeforeGameTickerCh <- true
		}()
		<-time.After(1 * time.Second)
		if err := mockedManager.ctx.Err(); err != nil {
			t.Fatal("expected context to be NOT canceled")
		}
		mockedManager.ctx.Done()
	})

	t.Run("should pass if game started", func(t *testing.T) {
		mockedManager := NewManager(cfg, workerPool, "TestOwner", "test123", 1, 3, 20, types.MockedQuestions)

		go mockedManager.BeforeGameLiquidationHandler()

		mockedManager.stopBeforeGameTickerCh <- true
		if mockedManager.gameState != 1 {
			log.Fatalln("expected game to be started")
		}
		mockedManager.ctx.Done()
	})
}

// CheckCorrectnessOfTime - function to check correctness of changing of time.
// It is used so many times to make sure that time handling correctly.
func CheckCorrectnessOfTime(t *testing.T, PreviousCurrentTime, CurrentTime int) {
	if PreviousCurrentTime == CurrentTime {
		t.Fatal("expected time to be different ")
	}
}
