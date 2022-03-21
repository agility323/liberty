package avatardata

/*
{
	"1": STATE(0/1/2),	// clean
	"2": STATE(0/1/2),	// rescue
}

clean: wait to clean - progress bar - reward
rescure: wait to rescue - progress bar - finish - reward

nil get finished rewarded
*/

const (
	InteractStateNil int = iota
	InteractStateGet
	InteractStateFinished
	InteractStateRewarded
)

// TODO: should replace with json config
var InitInteracts = []string {"1", }
var InteractJumpMap = map[string][]string {
	"1": []string {"2", },
	"2": []string {"3", },
}
var InteractRewardMap = map[string][][]int {
	"1": [][]int {[]int {1, 10}, },
	"2": [][]int {[]int {1, 20}, []int {2, 20}, },
	"3": [][]int {[]int{1, 15}, []int {2, 15}, []int {3, 15}, },
}

func MakeInitInteractData() map[string]int {
	m := make(map[string]int)
	for _, id := range InitInteracts {
		m[id] = InteractStateGet
	}
	return m
}
