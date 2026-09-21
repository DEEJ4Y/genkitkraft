package commands

import (
	"context"

	"github.com/DEEJ4Y/genkitkraft/internal/domain/agent"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/playground"
	"github.com/DEEJ4Y/genkitkraft/internal/domain/provider"
)

// fakePlaygroundRepo implements playgroundrepo.PlaygroundRepository. Only
// GetSession and GetLatestMessageBySession are ever configured/asserted on
// by the gap command tests that use it — every other method is a harmless
// no-op, per this repo's convention of full-interface stubs (see stubCache
// in genkit_chat_provider/builtin_tools_test.go).
type fakePlaygroundRepo struct {
	getSessionResult *playground.Session
	getSessionErr    error

	getLatestMessageResult *playground.Message
	getLatestMessageErr    error
	getLatestMessageCalls  int
}

func (f *fakePlaygroundRepo) CreateSession(context.Context, *playground.Session) error { return nil }

func (f *fakePlaygroundRepo) GetSession(_ context.Context, _ string) (*playground.Session, error) {
	return f.getSessionResult, f.getSessionErr
}

func (f *fakePlaygroundRepo) ListSessionsByAgent(context.Context, string) ([]*playground.Session, error) {
	return nil, nil
}

func (f *fakePlaygroundRepo) DeleteSession(context.Context, string) error { return nil }

func (f *fakePlaygroundRepo) UpdateSessionTitle(context.Context, string, string) error { return nil }

func (f *fakePlaygroundRepo) CreateMessage(context.Context, *playground.Message) error { return nil }

func (f *fakePlaygroundRepo) ListMessagesBySession(context.Context, string) ([]*playground.Message, error) {
	return nil, nil
}

func (f *fakePlaygroundRepo) CreateStreamingMessage(context.Context, string) (*playground.Message, error) {
	return nil, nil
}

func (f *fakePlaygroundRepo) AppendMessageChunk(context.Context, string, string) (int, error) {
	return 0, nil
}

func (f *fakePlaygroundRepo) GetMessageChunksSince(context.Context, string, int) ([]playground.MessageChunk, error) {
	return nil, nil
}

func (f *fakePlaygroundRepo) CompleteMessage(context.Context, string) error { return nil }

func (f *fakePlaygroundRepo) FailMessage(context.Context, string) error { return nil }

func (f *fakePlaygroundRepo) GetMessage(context.Context, string) (*playground.Message, error) {
	return nil, nil
}

func (f *fakePlaygroundRepo) GetLatestMessageBySession(_ context.Context, _ string) (*playground.Message, error) {
	f.getLatestMessageCalls++
	return f.getLatestMessageResult, f.getLatestMessageErr
}

// fakeAgentRepo implements agentrepo.AgentRepository; only GetByID matters
// to the gap dedup pipeline.
type fakeAgentRepo struct {
	getByIDResult *agent.Agent
	getByIDErr    error
}

func (f *fakeAgentRepo) List(context.Context, int, int) ([]*agent.Agent, error) { return nil, nil }

func (f *fakeAgentRepo) Count(context.Context) (int, error) { return 0, nil }

func (f *fakeAgentRepo) GetByID(_ context.Context, _ string) (*agent.Agent, error) {
	return f.getByIDResult, f.getByIDErr
}

func (f *fakeAgentRepo) Create(context.Context, *agent.Agent) error { return nil }

func (f *fakeAgentRepo) Update(context.Context, *agent.Agent) error { return nil }

func (f *fakeAgentRepo) Delete(context.Context, string) error { return nil }

// fakeProviderRepo implements providerrepo.ProviderRepository; only GetByID
// matters to the gap dedup pipeline.
type fakeProviderRepo struct {
	getByIDResult *provider.Provider
	getByIDErr    error
}

func (f *fakeProviderRepo) List(context.Context) ([]*provider.Provider, error) { return nil, nil }

func (f *fakeProviderRepo) GetByID(_ context.Context, _ string) (*provider.Provider, error) {
	return f.getByIDResult, f.getByIDErr
}

func (f *fakeProviderRepo) GetByType(context.Context, provider.ProviderType) (*provider.Provider, error) {
	return nil, nil
}

func (f *fakeProviderRepo) Create(context.Context, *provider.Provider) error { return nil }

func (f *fakeProviderRepo) Update(context.Context, *provider.Provider) error { return nil }

func (f *fakeProviderRepo) Delete(context.Context, string) error { return nil }

// fakeEncryptor implements encryptor.Encryptor, returning a fixed decrypted
// value regardless of ciphertext.
type fakeEncryptor struct {
	decrypted string
	err       error
}

func (f *fakeEncryptor) Encrypt(s string) (string, error) { return s, nil }

func (f *fakeEncryptor) Decrypt(string) (string, error) { return f.decrypted, f.err }
