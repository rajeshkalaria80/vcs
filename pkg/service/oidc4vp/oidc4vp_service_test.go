/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package oidc4vp_test

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trustbloc/did-go/doc/did"
	ldcontext "github.com/trustbloc/did-go/doc/ld/context"
	lddocloader "github.com/trustbloc/did-go/doc/ld/documentloader"
	util "github.com/trustbloc/did-go/doc/util/time"
	ariesmockstorage "github.com/trustbloc/did-go/legacy/mock/storage"
	vdrapi "github.com/trustbloc/did-go/vdr/api"
	vdrmock "github.com/trustbloc/did-go/vdr/mock"
	"github.com/trustbloc/kms-go/crypto/tinkcrypto"
	"github.com/trustbloc/kms-go/doc/jose/jwk"
	"github.com/trustbloc/kms-go/doc/util/fingerprint"
	"github.com/trustbloc/kms-go/doc/util/jwkkid"
	"github.com/trustbloc/kms-go/kms/localkms"
	mockkms "github.com/trustbloc/kms-go/mock/kms"
	"github.com/trustbloc/kms-go/secretlock/noop"
	ariescrypto "github.com/trustbloc/kms-go/spi/crypto"
	"github.com/trustbloc/kms-go/spi/kms"
	"github.com/trustbloc/vc-go/presexch"
	"github.com/trustbloc/vc-go/signature/suite"
	"github.com/trustbloc/vc-go/verifiable"

	"github.com/trustbloc/vcs/pkg/doc/vc"
	vcsverifiable "github.com/trustbloc/vcs/pkg/doc/verifiable"
	"github.com/trustbloc/vcs/pkg/event/spi"
	"github.com/trustbloc/vcs/pkg/internal/testutil"
	"github.com/trustbloc/vcs/pkg/kms/signer"
	profileapi "github.com/trustbloc/vcs/pkg/profile"
	"github.com/trustbloc/vcs/pkg/service/oidc4vp"
)

var (
	//go:embed testdata/university_degree.jsonld
	sampleVCJsonLD string
	//go:embed testdata/university_degree.jwt
	sampleVCJWT string
)

const (
	profileID      = "testProfileID"
	profileVersion = "v1.0"
)

func TestService_InitiateOidcInteraction(t *testing.T) {
	customKMS := createKMS(t)

	customCrypto, err := tinkcrypto.New()
	require.NoError(t, err)

	kmsRegistry := NewMockKMSRegistry(gomock.NewController(t))
	kmsRegistry.EXPECT().GetKeyManager(gomock.Any()).AnyTimes().Return(
		&mockVCSKeyManager{crypto: customCrypto, kms: customKMS}, nil)

	txManager := NewMockTransactionManager(gomock.NewController(t))
	txManager.EXPECT().CreateTx(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(&oidc4vp.Transaction{
		ID:                     "TxID1",
		ProfileID:              "test4",
		PresentationDefinition: &presexch.PresentationDefinition{},
	}, "nonce1", nil)
	requestObjectPublicStore := NewMockRequestObjectPublicStore(gomock.NewController(t))
	requestObjectPublicStore.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().DoAndReturn(func(ctx context.Context, token string, event *spi.Event) (string, error) {
		return "someurl/abc", nil
	})

	s := oidc4vp.NewService(&oidc4vp.Config{
		EventSvc:                 &mockEvent{},
		EventTopic:               spi.VerifierEventTopic,
		TransactionManager:       txManager,
		RequestObjectPublicStore: requestObjectPublicStore,
		KMSRegistry:              kmsRegistry,
		RedirectURL:              "test://redirect",
		TokenLifetime:            time.Second * 100,
	})

	keyID, _, err := customKMS.CreateAndExportPubKeyBytes(kms.ED25519Type)
	require.NoError(t, err)

	correctProfile := &profileapi.Verifier{
		ID:             "test1",
		Name:           "test2",
		URL:            "test3",
		Active:         true,
		OrganizationID: "test4",
		OIDCConfig: &profileapi.OIDC4VPConfig{
			KeyType: kms.ED25519Type,
		},
		Checks: &profileapi.VerificationChecks{
			Credential: profileapi.CredentialChecks{
				Proof: false,
				Format: []vcsverifiable.Format{
					vcsverifiable.Jwt,
				},
			},
			Presentation: &profileapi.PresentationChecks{
				Format: []vcsverifiable.Format{
					vcsverifiable.Jwt,
				},
			},
		},
		SigningDID: &profileapi.SigningDID{
			DID:      "did:test:acde",
			Creator:  "did:test:acde#" + keyID,
			KMSKeyID: keyID,
		},
	}

	t.Run("Success", func(t *testing.T) {
		info, err := s.InitiateOidcInteraction(context.TODO(), &presexch.PresentationDefinition{
			ID: "test",
		}, "test", correctProfile)

		require.NoError(t, err)
		require.NotNil(t, info)
	})

	t.Run("No signature did", func(t *testing.T) {
		incorrectProfile := &profileapi.Verifier{}
		require.NoError(t, copier.Copy(incorrectProfile, correctProfile))
		incorrectProfile.SigningDID = nil

		info, err := s.InitiateOidcInteraction(context.TODO(), &presexch.PresentationDefinition{}, "test", incorrectProfile)

		require.Error(t, err)
		require.Nil(t, info)
	})

	t.Run("Tx create failed", func(t *testing.T) {
		txManagerErr := NewMockTransactionManager(gomock.NewController(t))
		txManagerErr.EXPECT().CreateTx(
			gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil, "", errors.New("fail"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:                 &mockEvent{},
			EventTopic:               spi.VerifierEventTopic,
			TransactionManager:       txManagerErr,
			RequestObjectPublicStore: requestObjectPublicStore,
			KMSRegistry:              kmsRegistry,
			RedirectURL:              "test://redirect",
		})

		info, err := withError.InitiateOidcInteraction(
			context.TODO(),
			&presexch.PresentationDefinition{},
			"test",
			correctProfile,
		)

		require.Contains(t, err.Error(), "create oidc tx")
		require.Nil(t, info)
	})

	t.Run("publish request object failed", func(t *testing.T) {
		requestObjectPublicStoreErr := NewMockRequestObjectPublicStore(gomock.NewController(t))
		requestObjectPublicStoreErr.EXPECT().Publish(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().Return("", errors.New("fail"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:                 &mockEvent{},
			EventTopic:               spi.VerifierEventTopic,
			TransactionManager:       txManager,
			RequestObjectPublicStore: requestObjectPublicStoreErr,
			KMSRegistry:              kmsRegistry,
			RedirectURL:              "test://redirect",
		})

		info, err := withError.InitiateOidcInteraction(
			context.TODO(),
			&presexch.PresentationDefinition{},
			"test",
			correctProfile,
		)

		require.Contains(t, err.Error(), "publish request object")
		require.Nil(t, info)
	})

	t.Run("fail to get kms form registry", func(t *testing.T) {
		kmsRegistry := NewMockKMSRegistry(gomock.NewController(t))
		kmsRegistry.EXPECT().GetKeyManager(gomock.Any()).AnyTimes().Return(nil, errors.New("fail"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:                 &mockEvent{},
			EventTopic:               spi.VerifierEventTopic,
			TransactionManager:       txManager,
			RequestObjectPublicStore: requestObjectPublicStore,
			KMSRegistry:              kmsRegistry,
			RedirectURL:              "test://redirect",
		})

		info, err := withError.InitiateOidcInteraction(
			context.TODO(),
			&presexch.PresentationDefinition{},
			"test",
			correctProfile,
		)

		require.Contains(t, err.Error(), "get key manager")
		require.Nil(t, info)
	})

	t.Run("Invalid key", func(t *testing.T) {
		incorrectProfile := &profileapi.Verifier{}
		require.NoError(t, copier.Copy(incorrectProfile, correctProfile))
		incorrectProfile.SigningDID.KMSKeyID = "invalid"

		info, err := s.InitiateOidcInteraction(context.TODO(), &presexch.PresentationDefinition{}, "test", incorrectProfile)

		require.Error(t, err)
		require.Nil(t, info)
	})

	t.Run("Invalid key type", func(t *testing.T) {
		incorrectProfile := &profileapi.Verifier{}
		require.NoError(t, copier.Copy(incorrectProfile, correctProfile))
		incorrectProfile.OIDCConfig.KeyType = "invalid"

		info, err := s.InitiateOidcInteraction(context.TODO(), &presexch.PresentationDefinition{}, "test", incorrectProfile)

		require.Error(t, err)
		require.Nil(t, info)
	})
}

func TestService_VerifyOIDCVerifiablePresentation(t *testing.T) {
	keyManager := createKMS(t)

	crypto, err := tinkcrypto.New()
	require.NoError(t, err)

	txManager := NewMockTransactionManager(gomock.NewController(t))
	profileService := NewMockProfileService(gomock.NewController(t))
	presentationVerifier := NewMockPresentationVerifier(gomock.NewController(t))
	vp, pd, issuer, vdr, loader := newVPWithPD(t, keyManager, crypto)

	s := oidc4vp.NewService(&oidc4vp.Config{
		EventSvc:             &mockEvent{},
		EventTopic:           spi.VerifierEventTopic,
		TransactionManager:   txManager,
		PresentationVerifier: presentationVerifier,
		ProfileService:       profileService,
		DocumentLoader:       loader,
		VDR:                  vdr,
	})

	txManager.EXPECT().GetByOneTimeToken("nonce1").AnyTimes().Return(&oidc4vp.Transaction{
		ID:                     "txID1",
		ProfileID:              profileID,
		ProfileVersion:         profileVersion,
		PresentationDefinition: pd,
	}, true, nil)

	txManager.EXPECT().StoreReceivedClaims(oidc4vp.TxID("txID1"), gomock.Any()).AnyTimes().Return(nil)

	profileService.EXPECT().GetProfile(profileID, profileVersion).AnyTimes().Return(&profileapi.Verifier{
		ID:      profileID,
		Version: profileVersion,
		Active:  true,
		Checks: &profileapi.VerificationChecks{
			Presentation: &profileapi.PresentationChecks{
				VCSubject: true,
				Format: []vcsverifiable.Format{
					vcsverifiable.Jwt,
				},
			},
		},
	}, nil)

	presentationVerifier.EXPECT().VerifyPresentation(context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().Return(nil, nil)

	t.Run("Success", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  vp,
				SignerDIDID:   issuer,
				VpTokenFormat: vcsverifiable.Jwt,
			}})

		require.NoError(t, err)
	})

	t.Run("Unsupported vp token format", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  vp,
				SignerDIDID:   issuer,
				VpTokenFormat: vcsverifiable.Ldp,
			}})

		require.ErrorContains(t, err, "profile does not support ldp vp_token format")
	})

	t.Run("Success - two VP tokens (merged)", func(t *testing.T) {
		var descriptors []*presexch.InputDescriptor
		err := json.Unmarshal([]byte(twoInputDescriptors), &descriptors)
		require.NoError(t, err)

		defs := &presexch.PresentationDefinition{
			InputDescriptors: descriptors,
		}

		mergedPS := &presexch.PresentationSubmission{
			DescriptorMap: []*presexch.InputDescriptorMapping{
				{
					ID:   defs.InputDescriptors[0].ID,
					Path: "$[0]",
					PathNested: &presexch.InputDescriptorMapping{
						ID:   defs.InputDescriptors[0].ID,
						Path: "$.verifiableCredential[0]",
					},
				},
				{
					ID:   defs.InputDescriptors[1].ID,
					Path: "$[1]",
					PathNested: &presexch.InputDescriptorMapping{
						ID:   defs.InputDescriptors[1].ID,
						Path: "$.verifiableCredential[0]",
					},
				},
			},
		}

		testLoader := testutil.DocumentLoader(t)

		vp1, issuer1, vdr1 := newVPWithPS(t, keyManager, crypto, mergedPS, "PhDDegree")
		vp2, issuer2, vdr2 := newVPWithPS(t, keyManager, crypto, mergedPS, "BachelorDegree")

		combinedDIDResolver := &vdrmock.VDRegistry{
			ResolveFunc: func(didID string, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
				switch didID {
				case issuer1:
					return vdr1.Resolve(didID, opts...)
				case issuer2:
					return vdr2.Resolve(didID, opts...)
				}

				return nil, fmt.Errorf("unexpected issuer")
			}}

		txManager2 := NewMockTransactionManager(gomock.NewController(t))

		s2 := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   txManager2,
			PresentationVerifier: presentationVerifier,
			ProfileService:       profileService,
			DocumentLoader:       testLoader,
			VDR:                  combinedDIDResolver,
		})

		txManager2.EXPECT().GetByOneTimeToken("nonce1").AnyTimes().Return(&oidc4vp.Transaction{
			ID:                     "txID1",
			ProfileID:              profileID,
			ProfileVersion:         profileVersion,
			PresentationDefinition: defs,
		}, true, nil)

		txManager2.EXPECT().StoreReceivedClaims(oidc4vp.TxID("txID1"), gomock.Any()).AnyTimes().Return(nil)

		err = s2.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{
				{
					Nonce:         "nonce1",
					Presentation:  vp1,
					SignerDIDID:   issuer1,
					VpTokenFormat: vcsverifiable.Jwt,
				},
				{
					Nonce:         "nonce1",
					Presentation:  vp2,
					SignerDIDID:   issuer2,
					VpTokenFormat: vcsverifiable.Jwt,
				},
			})

		require.NoError(t, err)
	})

	t.Run("Error - Two VP tokens without presentation ID", func(t *testing.T) {
		var descriptors []*presexch.InputDescriptor
		err := json.Unmarshal([]byte(twoInputDescriptors), &descriptors)
		require.NoError(t, err)

		defs := &presexch.PresentationDefinition{
			InputDescriptors: descriptors,
		}

		mergedPS := &presexch.PresentationSubmission{
			DescriptorMap: []*presexch.InputDescriptorMapping{
				{
					ID:   defs.InputDescriptors[0].ID,
					Path: "$[0]",
					PathNested: &presexch.InputDescriptorMapping{
						ID:   defs.InputDescriptors[0].ID,
						Path: "$.verifiableCredential[0]",
					},
				},
				{
					ID:   defs.InputDescriptors[1].ID,
					Path: "$[1]",
					PathNested: &presexch.InputDescriptorMapping{
						ID:   defs.InputDescriptors[1].ID,
						Path: "$.verifiableCredential[0]",
					},
				},
			},
		}

		testLoader := testutil.DocumentLoader(t)

		vp1, issuer1, vdr1 := newVPWithPS(t, keyManager, crypto, mergedPS, "PhDDegree")
		vp2, issuer2, vdr2 := newVPWithPS(t, keyManager, crypto, mergedPS, "BachelorDegree")

		combinedDIDResolver := &vdrmock.VDRegistry{
			ResolveFunc: func(didID string, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
				switch didID {
				case issuer1:
					return vdr1.Resolve(didID, opts...)
				case issuer2:
					return vdr2.Resolve(didID, opts...)
				}

				return nil, fmt.Errorf("unexpected issuer")
			}}

		txManager2 := NewMockTransactionManager(gomock.NewController(t))

		s2 := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   txManager2,
			PresentationVerifier: presentationVerifier,
			ProfileService:       profileService,
			DocumentLoader:       testLoader,
			VDR:                  combinedDIDResolver,
		})

		txManager2.EXPECT().GetByOneTimeToken("nonce1").AnyTimes().Return(&oidc4vp.Transaction{
			ID:                     "txID1",
			ProfileID:              profileID,
			ProfileVersion:         profileVersion,
			PresentationDefinition: defs,
		}, true, nil)

		txManager2.EXPECT().StoreReceivedClaims(oidc4vp.TxID("txID1"), gomock.Any()).AnyTimes().Return(nil)

		vp1.ID = ""
		vp2.ID = ""

		err = s2.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{
				{
					Nonce:         "nonce1",
					Presentation:  vp1,
					SignerDIDID:   issuer1,
					VpTokenFormat: vcsverifiable.Jwt,
				},
				{
					Nonce:         "nonce1",
					Presentation:  vp2,
					SignerDIDID:   issuer2,
					VpTokenFormat: vcsverifiable.Jwt,
				},
			})

		require.Error(t, err)
		require.Contains(t, err.Error(), "duplicate presentation ID: ")
	})

	t.Run("Must have at least one token", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{})

		require.Error(t, err)
		require.Contains(t, err.Error(), "must have at least one token")
	})

	t.Run("VC subject is not much with vp signer", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  vp,
				SignerDIDID:   "did:example1:ebfeb1f712ebc6f1c276e12ec21",
				VpTokenFormat: vcsverifiable.Jwt,
			}})

		require.Contains(t, err.Error(), "does not match with vp signer")
	})

	t.Run("Invalid Nonce", func(t *testing.T) {
		errTxManager := NewMockTransactionManager(gomock.NewController(t))
		errTxManager.EXPECT().GetByOneTimeToken("nonce1").AnyTimes().
			Return(nil, false, errors.New("invalid nonce1"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   errTxManager,
			PresentationVerifier: presentationVerifier,
			ProfileService:       profileService,
			DocumentLoader:       loader,
		})

		err := withError.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:        "nonce1",
				Presentation: vp,
				SignerDIDID:  "did:example123:ebfeb1f712ebc6f1c276e12ec21",
			}})

		require.Contains(t, err.Error(), "invalid nonce1")
	})

	t.Run("Invalid Nonce 2", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID2",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:        "nonce1",
				Presentation: vp,
				SignerDIDID:  "did:example123:ebfeb1f712ebc6f1c276e12ec21",
			}})

		require.Contains(t, err.Error(), "invalid nonce")
	})

	t.Run("Invalid Nonce", func(t *testing.T) {
		errProfileService := NewMockProfileService(gomock.NewController(t))
		errProfileService.EXPECT().GetProfile(profileID, profileVersion).Times(1).Return(nil,
			errors.New("get profile error"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   txManager,
			PresentationVerifier: presentationVerifier,
			ProfileService:       errProfileService,
			DocumentLoader:       loader,
		})

		err := withError.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:        "nonce1",
				Presentation: vp,
				SignerDIDID:  "did:example123:ebfeb1f712ebc6f1c276e12ec21",
			}})

		require.Contains(t, err.Error(), "get profile error")
	})

	t.Run("verification failed", func(t *testing.T) {
		errPresentationVerifier := NewMockPresentationVerifier(gomock.NewController(t))
		errPresentationVerifier.EXPECT().VerifyPresentation(
			context.Background(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).
			Return(nil, errors.New("verification failed"))
		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   txManager,
			PresentationVerifier: errPresentationVerifier,
			ProfileService:       profileService,
			DocumentLoader:       loader,
			VDR:                  vdr,
		})

		err := withError.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  vp,
				SignerDIDID:   "did:example123:ebfeb1f712ebc6f1c276e12ec21",
				VpTokenFormat: vcsverifiable.Jwt,
			}})

		require.Contains(t, err.Error(), "verification failed")
	})

	t.Run("Match failed", func(t *testing.T) {
		err := s.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  &verifiable.Presentation{},
				VpTokenFormat: vcsverifiable.Jwt,
			}})
		require.Contains(t, err.Error(), "match:")
	})

	t.Run("Store error", func(t *testing.T) {
		errTxManager := NewMockTransactionManager(gomock.NewController(t))
		errTxManager.EXPECT().GetByOneTimeToken("nonce1").AnyTimes().Return(&oidc4vp.Transaction{
			ID:                     "txID1",
			ProfileID:              profileID,
			ProfileVersion:         profileVersion,
			PresentationDefinition: pd,
		}, true, nil)

		errTxManager.EXPECT().StoreReceivedClaims(oidc4vp.TxID("txID1"), gomock.Any()).
			Return(errors.New("store error"))

		withError := oidc4vp.NewService(&oidc4vp.Config{
			EventSvc:             &mockEvent{},
			EventTopic:           spi.VerifierEventTopic,
			TransactionManager:   errTxManager,
			PresentationVerifier: presentationVerifier,
			ProfileService:       profileService,
			DocumentLoader:       loader,
			VDR:                  vdr,
		})

		err := withError.VerifyOIDCVerifiablePresentation(context.Background(), "txID1",
			[]*oidc4vp.ProcessedVPToken{{
				Nonce:         "nonce1",
				Presentation:  vp,
				SignerDIDID:   issuer,
				VpTokenFormat: vcsverifiable.Jwt,
			}})

		require.Contains(t, err.Error(), "store error")
	})
}

func TestService_GetTx(t *testing.T) {
	txManager := NewMockTransactionManager(gomock.NewController(t))
	txManager.EXPECT().Get(oidc4vp.TxID("test")).Times(1).Return(&oidc4vp.Transaction{
		ProfileID: "testP1",
	}, nil)

	svc := oidc4vp.NewService(&oidc4vp.Config{
		TransactionManager: txManager,
	})

	t.Run("Success", func(t *testing.T) {
		tx, err := svc.GetTx(context.Background(), "test")
		require.NoError(t, err)
		require.NotNil(t, tx)
		require.Equal(t, "testP1", tx.ProfileID)
	})
}

func TestService_DeleteClaims(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		txManager := NewMockTransactionManager(gomock.NewController(t))
		txManager.EXPECT().DeleteReceivedClaims("claimsID").Times(1).Return(nil)

		svc := oidc4vp.NewService(&oidc4vp.Config{
			TransactionManager: txManager,
		})

		err := svc.DeleteClaims(context.Background(), "claimsID")
		require.NoError(t, err)
	})

	t.Run("Error", func(t *testing.T) {
		txManager := NewMockTransactionManager(gomock.NewController(t))
		txManager.EXPECT().DeleteReceivedClaims("claimsID").Times(1).Return(fmt.Errorf("delete error"))

		svc := oidc4vp.NewService(&oidc4vp.Config{
			TransactionManager: txManager,
		})

		err := svc.DeleteClaims(context.Background(), "claimsID")
		require.Error(t, err)
		require.Contains(t, err.Error(), "delete error")
	})
}

func TestService_RetrieveClaims(t *testing.T) {
	svc := oidc4vp.NewService(&oidc4vp.Config{})
	loader := testutil.DocumentLoader(t)

	t.Run("Success JWT", func(t *testing.T) {
		jwtvc, err := verifiable.ParseCredential([]byte(sampleVCJWT),
			verifiable.WithJSONLDDocumentLoader(loader),
			verifiable.WithDisabledProofCheck())

		require.NoError(t, err)

		claims := svc.RetrieveClaims(context.Background(), &oidc4vp.Transaction{
			ReceivedClaims: &oidc4vp.ReceivedClaims{Credentials: map[string]*verifiable.Credential{
				"id": jwtvc,
			}}})

		require.NotNil(t, claims)
		subjects, ok := claims["http://example.gov/credentials/3732"].SubjectData.([]verifiable.Subject)

		require.True(t, ok)
		require.Equal(t, "did:example:ebfeb1f712ebc6f1c276e12ec21", subjects[0].ID)

		require.NotEmpty(t, claims["http://example.gov/credentials/3732"].Issuer)
		require.NotEmpty(t, claims["http://example.gov/credentials/3732"].IssuanceDate)
		require.Empty(t, claims["http://example.gov/credentials/3732"].ExpirationDate)
	})

	t.Run("Success JsonLD", func(t *testing.T) {
		ldvc, err := verifiable.ParseCredential([]byte(sampleVCJsonLD),
			verifiable.WithJSONLDDocumentLoader(loader),
			verifiable.WithDisabledProofCheck())

		require.NoError(t, err)

		claims := svc.RetrieveClaims(context.Background(), &oidc4vp.Transaction{
			ReceivedClaims: &oidc4vp.ReceivedClaims{Credentials: map[string]*verifiable.Credential{
				"id": ldvc,
			}}})

		require.NotNil(t, claims)
		subjects, ok := claims["http://example.gov/credentials/3732"].SubjectData.([]verifiable.Subject)

		require.True(t, ok)
		require.Equal(t, "did:example:ebfeb1f712ebc6f1c276e12ec21", subjects[0].ID)

		require.NotEmpty(t, claims["http://example.gov/credentials/3732"].Issuer)
		require.NotEmpty(t, claims["http://example.gov/credentials/3732"].IssuanceDate)
		require.NotEmpty(t, claims["http://example.gov/credentials/3732"].ExpirationDate)
	})

	t.Run("Error", func(t *testing.T) {
		credential := &verifiable.Credential{
			JWT:          "abc",
			SDJWTHashAlg: "sha-256",
		}

		claims := svc.RetrieveClaims(context.Background(), &oidc4vp.Transaction{
			ReceivedClaims: &oidc4vp.ReceivedClaims{Credentials: map[string]*verifiable.Credential{
				"id": credential,
			}}})

		require.Empty(t, claims)
	})
}

func createKMS(t *testing.T) *localkms.LocalKMS {
	t.Helper()

	p, err := mockkms.NewProviderForKMS(ariesmockstorage.NewMockStoreProvider(), &noop.NoLock{})
	require.NoError(t, err)

	k, err := localkms.New("local-lock://custom/primary/key/", p)
	require.NoError(t, err)

	return k
}

type mockVCSKeyManager struct {
	crypto ariescrypto.Crypto
	kms    *localkms.LocalKMS
}

func (m *mockVCSKeyManager) NewVCSigner(creator string,
	signatureType vcsverifiable.SignatureType) (vc.SignerAlgorithm, error) {
	return signer.NewKMSSigner(m.kms, m.crypto, creator, signatureType, nil)
}

func (m *mockVCSKeyManager) SupportedKeyTypes() []kms.KeyType {
	return []kms.KeyType{kms.ED25519Type}
}
func (m *mockVCSKeyManager) CreateJWKKey(_ kms.KeyType) (string, *jwk.JWK, error) {
	return "", nil, nil
}
func (m *mockVCSKeyManager) CreateCryptoKey(_ kms.KeyType) (string, interface{}, error) {
	return "", nil, nil
}

type mockEvent struct {
	err error
}

func (m *mockEvent) Publish(_ context.Context, _ string, _ ...*spi.Event) error {
	if m.err != nil {
		return m.err
	}

	return nil
}

func newVPWithPD(t *testing.T, keyManager kms.KeyManager, crypto ariescrypto.Crypto) (
	*verifiable.Presentation, *presexch.PresentationDefinition, string,
	vdrapi.Registry, *lddocloader.DocumentLoader) {
	uri := randomURI()

	customType := "CustomType"

	expected, issuer, pubKeyFetcher := newSignedJWTVC(t, keyManager, crypto, []string{uri}, "", "")
	expected.Types = append(expected.Types, customType)

	defs := &presexch.PresentationDefinition{
		InputDescriptors: []*presexch.InputDescriptor{{
			ID: uuid.New().String(),
			Schema: []*presexch.Schema{{
				URI: fmt.Sprintf("%s#%s", uri, customType),
			}},
		}},
	}

	docLoader := createTestDocumentLoader(t, uri, customType)

	return newVP(t,
		&presexch.PresentationSubmission{DescriptorMap: []*presexch.InputDescriptorMapping{{
			ID:   defs.InputDescriptors[0].ID,
			Path: "$.verifiableCredential[0]",
		}}},
		expected,
	), defs, issuer, pubKeyFetcher, docLoader
}

func newVPWithPS(t *testing.T, keyManager kms.KeyManager, crypto ariescrypto.Crypto,
	ps *presexch.PresentationSubmission, value string) (
	*verifiable.Presentation, string, vdrapi.Registry) {
	expected, issuer, pubKeyFetcher := newSignedJWTVC(t, keyManager, crypto, nil, "degree", value)

	return newVP(t, ps,
		expected,
	), issuer, pubKeyFetcher
}

func newVP(t *testing.T, submission *presexch.PresentationSubmission,
	vcs ...*verifiable.Credential) *verifiable.Presentation {
	vp, err := verifiable.NewPresentation(verifiable.WithCredentials(vcs...))
	vp.ID = uuid.New().String() // TODO: Can we rely on this for code
	require.NoError(t, err)

	vp.Context = append(vp.Context, "https://identity.foundation/presentation-exchange/submission/v1")
	vp.Type = append(vp.Type, "PresentationSubmission")

	if submission != nil {
		vp.CustomFields = make(map[string]interface{})
		vp.CustomFields["presentation_submission"] = toMap(t, submission)
	}

	return vp
}

func newVC(issuer string, ctx []string) *verifiable.Credential {
	cred := &verifiable.Credential{
		Context: []string{verifiable.ContextURI},
		Types:   []string{verifiable.VCType},
		ID:      "http://test.credential.com/123",
		Issuer:  verifiable.Issuer{ID: issuer},
		Issued: &util.TimeWrapper{
			Time: time.Now(),
		},
		Subject: map[string]interface{}{
			"id": issuer,
		},
	}

	if ctx != nil {
		cred.Context = append(cred.Context, ctx...)
	}

	return cred
}

func newDegreeVC(issuer string, degreeType string, ctx []string) *verifiable.Credential {
	cred := &verifiable.Credential{
		Context: []string{verifiable.ContextURI},
		Types:   []string{verifiable.VCType},
		ID:      uuid.New().String(),
		Issuer:  verifiable.Issuer{ID: issuer},
		Issued: &util.TimeWrapper{
			Time: time.Now(),
		},
		Subject: map[string]interface{}{
			"id": issuer,
			"degree": map[string]interface{}{
				"type":   degreeType,
				"degree": "MIT",
			},
		},
	}

	if ctx != nil {
		cred.Context = append(cred.Context, ctx...)
	}

	return cred
}

func newSignedJWTVC(t *testing.T,
	keyManager kms.KeyManager, crypto ariescrypto.Crypto, ctx []string,
	vcType string, value string) (*verifiable.Credential, string, vdrapi.Registry) {
	t.Helper()

	keyID, kh, err := keyManager.Create(kms.ED25519Type)
	require.NoError(t, err)

	signer := suite.NewCryptoSigner(crypto, kh)

	pubKey, kt, err := keyManager.ExportPubKeyBytes(keyID)
	require.NoError(t, err)
	require.Equal(t, kms.ED25519Type, kt)

	key, err := jwkkid.BuildJWK(pubKey, kms.ED25519)
	require.NoError(t, err)

	issuer, verMethod := fingerprint.CreateDIDKeyByCode(fingerprint.ED25519PubKeyMultiCodec, pubKey)

	verificationMethod, err := did.NewVerificationMethodFromJWK(verMethod, "JsonWebKey2020", issuer, key)
	require.NoError(t, err)

	didResolver := &vdrmock.VDRegistry{
		ResolveFunc: func(didID string, opts ...vdrapi.DIDMethodOption) (*did.DocResolution, error) {
			return makeMockDIDResolution(issuer, verificationMethod, did.Authentication), nil
		}}

	var vc *verifiable.Credential

	switch vcType {
	case "degree":
		vc = newDegreeVC(issuer, value, ctx)
	default:
		vc = newVC(issuer, ctx)
	}

	vc.Issuer = verifiable.Issuer{ID: issuer}

	claims, err := vc.JWTClaims(false)
	require.NoError(t, err)

	jwsAlgo, err := verifiable.KeyTypeToJWSAlgo(kms.ED25519Type)
	require.NoError(t, err)

	jws, err := claims.MarshalJWS(jwsAlgo, signer, verMethod)
	require.NoError(t, err)

	vc.JWT = jws

	return vc, issuer, didResolver
}

func makeMockDIDResolution(id string, vm *did.VerificationMethod, vr did.VerificationRelationship) *did.DocResolution {
	ver := []did.Verification{{
		VerificationMethod: *vm,
		Relationship:       vr,
	}}

	doc := &did.Doc{
		ID: id,
	}

	switch vr { //nolint:exhaustive
	case did.VerificationRelationshipGeneral:
		doc.VerificationMethod = []did.VerificationMethod{*vm}
	case did.Authentication:
		doc.Authentication = ver
	case did.AssertionMethod:
		doc.AssertionMethod = ver
	}

	return &did.DocResolution{
		DIDDocument: doc,
	}
}

func randomURI() string {
	return fmt.Sprintf("https://my.test.context.jsonld/%s", uuid.New().String())
}

func createTestDocumentLoader(t *testing.T, contextURL string, types ...string) *lddocloader.DocumentLoader {
	include := fmt.Sprintf(`"ctx":"%s#"`, contextURL)

	for _, typ := range types {
		include += fmt.Sprintf(`,"%s":"ctx:%s"`, typ, typ)
	}

	jsonLDContext := fmt.Sprintf(`{
    "@context":{
      "@version":1.1,
      "@protected":true,
      "name":"http://schema.org/name",
      "ex":"https://example.org/examples#",
      "xsd":"http://www.w3.org/2001/XMLSchema#",
	  %s
	   }
	}`, include)

	loader := testutil.DocumentLoader(t, ldcontext.Document{
		URL:     contextURL,
		Content: []byte(jsonLDContext),
	})

	return loader
}

func toMap(t *testing.T, v interface{}) map[string]interface{} {
	bits, err := json.Marshal(v)
	require.NoError(t, err)

	m := make(map[string]interface{})

	err = json.Unmarshal(bits, &m)
	require.NoError(t, err)

	return m
}

const twoInputDescriptors = `
[
  {
    "id": "phd-degree",
    "name": "phd-degree",
    "purpose": "We can only hire with PhD degree.",
    "schema": [
      {
        "uri": "https://www.w3.org/2018/credentials#VerifiableCredential"
      }
    ],
    "constraints": {
      "fields": [
        {
          "path": [
            "$.credentialSubject.degree.type",
            "$.vc.credentialSubject.degree.type"
          ],
          "purpose": "We can only hire with PhD degree.",
          "filter": {
            "type": "string",
            "const": "PhDDegree"
          }
        }
      ]
    }
  },
  {
    "id": "bachelor-degree",
    "name": "bachelor-degree",
    "purpose": "We can only hire with bachelor degree.",
    "schema": [
      {
        "uri": "https://www.w3.org/2018/credentials#VerifiableCredential"
      }
    ],
    "constraints": {
      "fields": [
        {
          "path": [
            "$.credentialSubject.degree.type",
            "$.vc.credentialSubject.degree.type"
          ],
          "purpose": "We can only hire with bachelor degree.",
          "filter": {
            "type": "string",
            "const": "BachelorDegree"
          }
        }
      ]
    }
  }
]`

func Test_GetSupportedVPFormats(t *testing.T) {
	type args struct {
		kmsSupportedKeyTypes []kms.KeyType
		supportedVPFormats   []vcsverifiable.Format
		supportedVCFormats   []vcsverifiable.Format
	}
	tests := []struct {
		name string
		args args
		want *presexch.Format
	}{
		{
			name: "OK with duplications",
			args: args{
				kmsSupportedKeyTypes: []kms.KeyType{
					kms.ED25519Type,
					kms.ECDSAP256TypeDER,
				},
				supportedVPFormats: []vcsverifiable.Format{
					vcsverifiable.Jwt,
					vcsverifiable.Ldp,
				},
				supportedVCFormats: []vcsverifiable.Format{
					vcsverifiable.Jwt,
					vcsverifiable.Ldp,
				},
			},
			want: &presexch.Format{
				JwtVC: &presexch.JwtType{Alg: []string{
					"EdDSA",
					"ES256",
				}},
				JwtVP: &presexch.JwtType{Alg: []string{
					"EdDSA",
					"ES256",
				}},
				LdpVC: &presexch.LdpType{ProofType: []string{
					"Ed25519Signature2018",
					"Ed25519Signature2020",
					"JsonWebSignature2020",
				}},
				LdpVP: &presexch.LdpType{ProofType: []string{
					"Ed25519Signature2018",
					"Ed25519Signature2020",
					"JsonWebSignature2020",
				}},
			},
		},
		{
			name: "OK",
			args: args{
				kmsSupportedKeyTypes: []kms.KeyType{
					kms.ED25519Type,
					kms.ECDSAP256TypeDER,
				},
				supportedVPFormats: []vcsverifiable.Format{
					vcsverifiable.Jwt,
				},
				supportedVCFormats: []vcsverifiable.Format{
					vcsverifiable.Ldp,
				},
			},
			want: &presexch.Format{
				JwtVC: nil,
				JwtVP: &presexch.JwtType{Alg: []string{
					"EdDSA",
					"ES256",
				}},
				LdpVC: &presexch.LdpType{ProofType: []string{
					"Ed25519Signature2018",
					"Ed25519Signature2020",
					"JsonWebSignature2020",
				}},
				LdpVP: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := oidc4vp.GetSupportedVPFormats(
				tt.args.kmsSupportedKeyTypes, tt.args.supportedVPFormats, tt.args.supportedVCFormats)

			assert.Equal(t, tt.want.JwtVC == nil, got.JwtVC == nil)
			if got.JwtVC != nil {
				assert.ElementsMatch(t, tt.want.JwtVC.Alg, got.JwtVC.Alg)
			}

			assert.Equal(t, tt.want.JwtVP == nil, got.JwtVP == nil)
			if got.JwtVC != nil {
				assert.ElementsMatch(t, tt.want.JwtVP.Alg, got.JwtVP.Alg)
			}

			assert.Equal(t, tt.want.LdpVC == nil, got.LdpVC == nil)
			if got.JwtVC != nil {
				assert.ElementsMatch(t, tt.want.LdpVC.ProofType, got.LdpVC.ProofType)
			}

			assert.Equal(t, tt.want.LdpVP == nil, got.LdpVP == nil)
			if got.JwtVC != nil {
				assert.ElementsMatch(t, tt.want.LdpVP.ProofType, got.LdpVP.ProofType)
			}
		})
	}
}
