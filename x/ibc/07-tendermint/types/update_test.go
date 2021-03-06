package types_test

import (
	"time"

	tmtypes "github.com/tendermint/tendermint/types"

	types "github.com/cosmos/cosmos-sdk/x/ibc/07-tendermint/types"
	commitmenttypes "github.com/cosmos/cosmos-sdk/x/ibc/23-commitment/types"
)

func (suite *TendermintTestSuite) TestCheckHeaderAndUpdateState() {
	var (
		clientState    *types.ClientState
		consensusState *types.ConsensusState
		newHeader      *types.Header
		currentTime    time.Time
	)

	// Setup different validators and signers for testing different types of updates
	altPrivVal := tmtypes.NewMockPV()
	altPubKey, err := altPrivVal.GetPubKey()
	suite.Require().NoError(err)

	altVal := tmtypes.NewValidator(altPubKey, height)

	// Create bothValSet with both suite validator and altVal. Would be valid update
	bothValSet := tmtypes.NewValidatorSet(append(suite.valSet.Validators, altVal))
	// Create alternative validator set with only altVal, invalid update (too much change in valSet)
	altValSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{altVal})

	signers := []tmtypes.PrivValidator{suite.privVal}

	// Create signer array and ensure it is in same order as bothValSet
	_, suiteVal := suite.valSet.GetByIndex(0)
	bothSigners := types.CreateSortedSignerArray(altPrivVal, suite.privVal, altVal, suiteVal)

	altSigners := []tmtypes.PrivValidator{altPrivVal}

	testCases := []struct {
		name    string
		setup   func()
		expPass bool
	}{
		{
			name: "successful update with next height and same validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with future height and different validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+5, height, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update with next height and different validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, bothValSet.Hash())
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "successful update for a previous height",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height-3, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height-1, height-3, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: true,
		},
		{
			name: "unsuccessful update with next height: update header mismatches nextValSetHash",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, bothValSet, suite.valSet, bothSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with next height: update header mismatches different nextValSetHash",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, bothValSet.Hash())
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, suite.valSet, bothValSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update with future height: too much change in validator set",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+5, height, suite.headerTime, altValSet, suite.valSet, altSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful updates, passed in incorrect trusted validators for given consensus state",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+5, height, suite.headerTime, bothValSet, bothValSet, bothSigners)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: trusting period has passed since last client timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				// make current time pass trusting period from last timestamp on clientstate
				currentTime = suite.now.Add(trustingPeriod)
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is past current timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.now.Add(time.Minute), suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "unsuccessful update: header timestamp is not past last client timestamp",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.clientTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header basic validation failed",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				newHeader = types.CreateTestHeader(chainID, height+1, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				// cause new header to fail validatebasic by changing commit height to mismatch header height
				newHeader.SignedHeader.Commit.Height = height - 1
				currentTime = suite.now
			},
			expPass: false,
		},
		{
			name: "header height < consensus height",
			setup: func() {
				clientState = types.NewClientState(chainID, types.DefaultTrustLevel, trustingPeriod, ubdPeriod, maxClockDrift, height+5, commitmenttypes.GetSDKSpecs())
				consensusState = types.NewConsensusState(suite.clientTime, commitmenttypes.NewMerkleRoot(suite.header.Header.GetAppHash()), height, suite.valsHash)
				// Make new header at height less than latest client state
				newHeader = types.CreateTestHeader(chainID, height-1, height, suite.headerTime, suite.valSet, suite.valSet, signers)
				currentTime = suite.now
			},
			expPass: false,
		},
	}

	for i, tc := range testCases {
		tc := tc
		// setup test
		tc.setup()

		// Set current timestamp in context
		ctx := suite.chainA.GetContext().WithBlockTime(currentTime)

		// Set trusted consensus state in client store
		suite.chainA.App.IBCKeeper.ClientKeeper.SetClientConsensusState(ctx, clientID, consensusState.GetHeight(), consensusState)

		expectedConsensus := &types.ConsensusState{
			Height:             uint64(newHeader.GetHeight()),
			Timestamp:          newHeader.GetTime(),
			Root:               commitmenttypes.NewMerkleRoot(newHeader.Header.GetAppHash()),
			NextValidatorsHash: newHeader.Header.NextValidatorsHash,
		}

		newClientState, consensusState, err := clientState.CheckHeaderAndUpdateState(
			ctx,
			suite.cdc,
			suite.chainA.App.IBCKeeper.ClientKeeper.ClientStore(suite.chainA.GetContext(), clientID), // pass in clientID prefixed clientStore
			newHeader,
		)

		if tc.expPass {
			suite.Require().NoError(err, "valid test case %d failed: %s", i, tc.name)

			// Determine if clientState should be updated or not
			if uint64(newHeader.GetHeight()) > clientState.LatestHeight {
				// Header Height is greater than clientState latest Height, clientState should be updated with header.GetHeight()
				suite.Require().Equal(uint64(newHeader.GetHeight()), newClientState.GetLatestHeight(), "clientstate height did not update")
			} else {
				// Update will add past consensus state, clientState should not be updated at all
				suite.Require().Equal(clientState.LatestHeight, newClientState.GetLatestHeight(), "client state height updated for past header")
			}

			suite.Require().Equal(expectedConsensus, consensusState, "valid test case %d failed: %s", i, tc.name)
		} else {
			suite.Require().Error(err, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(newClientState, "invalid test case %d passed: %s", i, tc.name)
			suite.Require().Nil(consensusState, "invalid test case %d passed: %s", i, tc.name)
		}
	}
}
