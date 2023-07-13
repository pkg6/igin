package firebase

import (
	"context"
	"fmt"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/pkg6/go-requests"
	"google.golang.org/api/option"
)

type AuthClient struct {
	ApiKey string
	Ctx    context.Context
	*auth.Client
}

func newFirebaseApp(ctx context.Context, projectID, credentialsFile string) (*firebase.App, error) {
	config := &firebase.Config{ProjectID: projectID}
	app, err := firebase.NewApp(ctx, config, option.WithCredentialsFile(credentialsFile))
	if err != nil {
		return nil, err
	}
	return app, nil
}

// NewAuthClient
//projectID && apiKey https://console.firebase.google.com/project/xxxxxx/settings/general?hl=zh-cn
//credentialsFile https://console.firebase.google.com/project/xxxxxx/settings/serviceaccounts/adminsdk?hl=zh-cn
func NewAuthClient(projectID, credentialsFile, apiKey string) (*AuthClient, error) {
	f := &AuthClient{ApiKey: apiKey, Ctx: context.Background()}
	firebaseApp, err := newFirebaseApp(f.Ctx, projectID, credentialsFile)
	if err != nil {
		return nil, err
	}
	f.Client, err = firebaseApp.Auth(f.Ctx)
	if err != nil {
		return nil, err
	}
	return f, nil
}

type IDToken struct {
	IDToken      string `json:"idToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    string `json:"expiresIn"`
}

// TokenToIDToken https://firebase.google.com/docs/reference/rest/auth?hl=zh-cn
func (f AuthClient) TokenToIDToken(token string) (IDToken, error) {
	var idToken IDToken
	data := struct {
		Token             string `json:"token"`
		ReturnSecureToken bool   `json:"returnSecureToken"`
	}{
		token, true,
	}
	resp, err := requests.PostJson(
		fmt.Sprintf(
			"https://identitytoolkit.googleapis.com/v1/accounts:signInWithCustomToken?key=%s",
			f.ApiKey,
		),
		data,
	)
	if err != nil {
		return idToken, err
	}
	if err := resp.Unmarshal(&idToken); err != nil {
		return idToken, err
	}
	return idToken, nil
}
