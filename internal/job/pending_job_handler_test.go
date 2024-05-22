package job

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
)

func TestRemoveUUID(t *testing.T) {
	uuid1 := uuid.New()
	uuid2 := uuid.New()
	uuid3 := uuid.New()

	tests := []struct {
		name   string
		uuids  []uuid.UUID
		target uuid.UUID
		want   []uuid.UUID
	}{
		{
			name:   "Remove from middle",
			uuids:  []uuid.UUID{uuid1, uuid2, uuid3},
			target: uuid2,
			want:   []uuid.UUID{uuid1, uuid3},
		},
		{
			name:   "Remove from end",
			uuids:  []uuid.UUID{uuid1, uuid2, uuid3},
			target: uuid3,
			want:   []uuid.UUID{uuid1, uuid2},
		},
		{
			name:   "Remove from beginning",
			uuids:  []uuid.UUID{uuid1, uuid2, uuid3},
			target: uuid1,
			want:   []uuid.UUID{uuid2, uuid3},
		},
		{
			name:   "Target not in slice",
			uuids:  []uuid.UUID{uuid1, uuid2, uuid3},
			target: uuid.New(),
			want:   []uuid.UUID{uuid1, uuid2, uuid3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := removeUUID(tt.uuids, tt.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("removeUUID() = %v, want %v", got, tt.want)
			}
		})
	}
}
