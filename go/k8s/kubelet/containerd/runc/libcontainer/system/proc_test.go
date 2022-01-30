package system

import (
	"fmt"
	"k8s.io/klog/v2"
	"testing"
)

func TestParseStartTime(t *testing.T) {
	data := map[string]Stat_t{
		"4902 (gunicorn: maste) S 4885 4902 4902 0 -1 4194560 29683 29929 61 83 78 16 96 17 20 0 1 0 9126532 52965376 1903 18446744073709551615 4194304 7461796 140733928751520 140733928698072 139816984959091 0 0 16781312 137447943 1 0 0 17 3 0 0 9 0 0 9559488 10071156 33050624 140733928758775 140733928758945 140733928758945 140733928759264 0": {
			PID:       4902,
			Name:      "gunicorn: maste",
			State:     'S',
			StartTime: 9126532,
		},
		"9534 (cat) R 9323 9534 9323 34828 9534 4194304 95 0 0 0 0 0 0 0 20 0 1 0 9214966 7626752 168 18446744073709551615 4194304 4240332 140732237651568 140732237650920 140570710391216 0 0 0 0 0 0 0 17 1 0 0 0 0 0 6340112 6341364 21553152 140732237653865 140732237653885 140732237653885 140732237656047 0": {
			PID:       9534,
			Name:      "cat",
			State:     'R',
			StartTime: 9214966,
		},

		"24767 (irq/44-mei_me) S 2 0 0 0 -1 2129984 0 0 0 0 0 0 0 0 -51 0 1 0 8722075 0 0 18446744073709551615 0 0 0 0 0 0 0 2147483647 0 0 0 0 17 1 50 1 0 0 0 0 0 0 0 0 0 0 0": {
			PID:       24767,
			Name:      "irq/44-mei_me",
			State:     'S',
			StartTime: 8722075,
		},
	}
	for line, expected := range data {
		stat, err := parseStat(line)
		if err != nil {
			t.Fatal(err)
		}

		klog.Info(fmt.Sprintf("PID:%d, Name:%s, State:%s, StartTime:%d", stat.PID, stat.Name, stat.State.String(), stat.StartTime))

		if stat.PID != expected.PID {
			t.Fatalf("expected PID %q but received %q", expected.PID, stat.PID)
		}
		if stat.State != expected.State {
			t.Fatalf("expected state %q but received %q", expected.State, stat.State)
		}
		if stat.Name != expected.Name {
			t.Fatalf("expected name %q but received %q", expected.Name, stat.Name)
		}
		if stat.StartTime != expected.StartTime {
			t.Fatalf("expected start time %q but received %q", expected.StartTime, stat.StartTime)
		}
	}

	klog.Info("current process: ")
	currentProcess := "4662 (bash) S 4659 4662 4454 34816 19057 4194560 3795 5137 0 0 7 1 1 3 20 0 1 0 377601972 125980672 1641 18446744073709551615 4194304 6830452 140732816467552 0 0 0 65536 3686404 1266761467 0 0 0 17 14 0 0 0 0 0 8928752 8971664 36052992 140732816472040 140732816472045 140732816472045 140732816474094 0"
	stat, err := parseStat(currentProcess)
	if err != nil {
		t.Fatal(err)
	}
	klog.Info(fmt.Sprintf("PID:%d, Name:%s, State:%s, StartTime:%d", stat.PID, stat.Name, stat.State.String(), stat.StartTime))
}
