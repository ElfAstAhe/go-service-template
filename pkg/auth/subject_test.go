package auth

import (
	"encoding/json"
	"testing"

	"github.com/ElfAstAhe/go-service-template/pkg/errs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubjectType(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		assert.True(t, SubjectUser.IsValid())
		assert.True(t, SubjectService.IsValid())
		assert.True(t, SubjectGuest.IsValid())
		assert.False(t, SubjectType("unknown").IsValid())
		assert.False(t, SubjectType("").IsValid())
	})

	t.Run("ParseSubjectType", func(t *testing.T) {
		tests := []struct {
			name      string
			input     string
			expected  SubjectType
			expectErr bool
		}{
			{
				name:      "valid user",
				input:     "user",
				expected:  SubjectUser,
				expectErr: false,
			},
			{
				name:      "valid service",
				input:     "service",
				expected:  SubjectService,
				expectErr: false,
			},
			{
				name:      "valid guest",
				input:     "guest",
				expected:  SubjectGuest,
				expectErr: false,
			},
			{
				name:      "invalid type",
				input:     "admin",
				expectErr: true,
			},
			{
				name:      "empty string",
				input:     "",
				expectErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				res, err := ParseSubjectType(tt.input)
				if tt.expectErr {
					var invalidArgErr *errs.InvalidArgumentError
					require.Error(t, err)
					assert.ErrorAs(t, err, &invalidArgErr)
					assert.Empty(t, res)
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.expected, res)
			})
		}
	})

	t.Run("JSON Marshal and Unmarshal", func(t *testing.T) {
		type payload struct {
			Type SubjectType `json:"type"`
		}

		// Валидный маршалинг и анмаршалинг
		raw, err := json.Marshal(payload{Type: SubjectService})
		require.NoError(t, err)
		assert.JSONEq(t, `{"type":"service"}`, string(raw))

		var decoded payload
		err = json.Unmarshal(raw, &decoded)
		require.NoError(t, err)
		assert.Equal(t, SubjectService, decoded.Type)

		// Ошибка анмаршалинга при невалидном типе
		var invalidDecoded payload
		err = json.Unmarshal([]byte(`{"type":"hacker"}`), &invalidDecoded)
		var invalidArgErr *errs.InvalidArgumentError
		require.Error(t, err)
		assert.ErrorAs(t, err, &invalidArgErr)
	})
}

func TestSubject(t *testing.T) {
	t.Run("NewSubject isolates metadata map", func(t *testing.T) {
		sourceMeta := map[string]string{
			"ip":         "127.0.0.1",
			"user-agent": "Mozilla/5.0",
		}
		roles := []string{"reader", "writer", "reader"}

		subj := NewSubject("id-1", "ivan", SubjectUser, roles, sourceMeta)

		// Мутируем исходную мапу после создания Subject
		sourceMeta["ip"] = "10.0.0.1"
		delete(sourceMeta, "user-agent")

		require.NotNil(t, subj)
		assert.Equal(t, "id-1", subj.ID)
		assert.Equal(t, "ivan", subj.Name)
		assert.Equal(t, SubjectUser, subj.Type)
		assert.Len(t, subj.Roles, 2)
		assert.Equal(t, "127.0.0.1", subj.Metadata["ip"])
		assert.Equal(t, "Mozilla/5.0", subj.Metadata["user-agent"])
	})

	t.Run("HasRole", func(t *testing.T) {
		subj := NewSubject("id-1", "ivan", SubjectUser, []string{"audit:read"}, nil)

		assert.True(t, subj.HasRole("audit:read"))
		assert.False(t, subj.HasRole("audit:write"))
		assert.False(t, subj.HasRole(""))
	})

	t.Run("Type predicates and IsAuthenticated", func(t *testing.T) {
		user := NewSubject("u-1", "ivan", SubjectUser, nil, nil)
		assert.True(t, user.IsUser())
		assert.False(t, user.IsService())
		assert.False(t, user.IsGuest())
		assert.True(t, user.IsAuthenticated())

		srv := NewSubject("s-1", "worker", SubjectService, nil, nil)
		assert.False(t, srv.IsUser())
		assert.True(t, srv.IsService())
		assert.False(t, srv.IsGuest())
		assert.True(t, srv.IsAuthenticated())

		guest := NewSubject("", "guest", SubjectGuest, nil, nil)
		assert.False(t, guest.IsUser())
		assert.False(t, guest.IsService())
		assert.True(t, guest.IsGuest())
		assert.False(t, guest.IsAuthenticated())

		userWithoutID := NewSubject("", "ivan", SubjectUser, nil, nil)
		assert.False(t, userWithoutID.IsAuthenticated())
	})

	t.Run("IsValid and String", func(t *testing.T) {
		valid := NewSubject("u-1", "ivan", SubjectUser, nil, nil)
		assert.True(t, valid.IsValid())
		assert.Equal(t, "u-1@ivan", valid.String())

		invalidNoID := NewSubject("", "ivan", SubjectUser, nil, nil)
		assert.False(t, invalidNoID.IsValid())

		invalidNoName := NewSubject("u-1", "", SubjectUser, nil, nil)
		assert.False(t, invalidNoName.IsValid())

		invalidNoType := NewSubject("u-1", "ivan", "", nil, nil)
		assert.False(t, invalidNoType.IsValid())
	})
}
