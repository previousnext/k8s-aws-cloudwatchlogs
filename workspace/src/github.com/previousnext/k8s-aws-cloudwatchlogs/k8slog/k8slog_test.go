package k8slog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	namespace, pod, container, err := metadata("staging-3869834092-dz5jn_previousnext-pnx-redmine3_app-xxxxxxxxx.log")
	assert.Nil(t, err)
	assert.Equal(t, "previousnext-pnx-redmine3", namespace)
	assert.Equal(t, "staging-3869834092-dz5jn", pod)
	assert.Equal(t, "app", container)
}
