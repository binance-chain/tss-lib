// Copyright © 2019 Binance
//
// This file is part of Binance. The full Binance copyright notice, including
// terms governing use, modification, and redistribution, is contained in the
// file LICENSE at the root of the source code distribution tree.

package signing

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"runtime"
	"sort"
	"sync/atomic"
	"testing"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/golang/protobuf/ptypes"
	"github.com/ipfs/go-log"
	"github.com/stretchr/testify/assert"

	"github.com/binance-chain/tss-lib/common"
	"github.com/binance-chain/tss-lib/crypto"
	"github.com/binance-chain/tss-lib/crypto/mta"
	"github.com/binance-chain/tss-lib/ecdsa/keygen"
	"github.com/binance-chain/tss-lib/test"
	"github.com/binance-chain/tss-lib/tss"
)

const (
	testParticipants = test.TestParticipants
	testThreshold    = test.TestThreshold
)

const (
	testFixtureDirFormat  = "%s/../../test/_ecdsa_fixtures"
	testFixtureFileFormat = "oneround_data_%d.json"
)

func setUp(level string) {
	if err := log.SetLogLevel("tss-lib", level); err != nil {
		panic(err)
	}
}

type signatureDataWithPartyId struct {
	signData *SignatureData
	partyId  *tss.PartyID
}

func TestE2EConcurrent(t *testing.T) {
	setUp("info")
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endCh := make(chan *SignatureData, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	msg := common.GetRandomPrimeInt(256)
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		P := NewLocalParty(msg, params, keys[i], outCh, endCh).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	var ended int32
signing:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			common.Logger.Errorf("Error: %s", err)
			assert.FailNow(t, err.Error())
			break signing

		case msg := <-outCh:
			for _, P := range parties {
				if P.PartyID().Index == msg.GetFrom().Index {
					continue
				}
				go updater(P, msg, errCh)
			}

		case data := <-endCh:
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Logf("Done. Received signature data from %d participants %+v", ended, data)

				// bigR is stored as bytes for the OneRoundData protobuf struct
				bigRX, bigRY := new(big.Int).SetBytes(parties[0].temp.BigR.GetX()), new(big.Int).SetBytes(parties[0].temp.BigR.GetY())
				bigR := crypto.NewECPointNoCurveCheck(tss.EC(), bigRX, bigRY)

				r := parties[0].temp.rI.X()
				fmt.Printf("sign result: R(%s, %s), r=%s\n", bigR.X().String(), bigR.Y().String(), r.String())

				modN := common.ModInt(tss.EC().Params().N)

				// BEGIN check s correctness
				sumS := big.NewInt(0)
				for _, p := range parties {
					sumS = modN.Add(sumS, p.temp.sI)
				}
				fmt.Printf("S: %s\n", sumS.String())
				// END check s correctness

				// BEGIN ECDSA verify
				pkX, pkY := keys[0].ECDSAPub.X(), keys[0].ECDSAPub.Y()
				pk := ecdsa.PublicKey{
					Curve: tss.EC(),
					X:     pkX,
					Y:     pkY,
				}
				ok := ecdsa.Verify(&pk, msg.Bytes(), bigR.X(), sumS)
				assert.True(t, ok, "ecdsa verify must pass")

				btcecSig := &btcec.Signature{R: r, S: sumS}
				btcecSig.Verify(msg.Bytes(), (*btcec.PublicKey)(&pk))
				assert.True(t, ok, "ecdsa verify 2 must pass")

				t.Log("ECDSA signing test done.")
				// END ECDSA verify

				break signing
			}
		}
	}
}

func TestE2EConcurrentAbortInRound2(t *testing.T) {
	setUp("info")
	attackers := []int{1, 2}
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endCh := make(chan *SignatureData, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	msg := common.GetRandomPrimeInt(256)
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		P := NewLocalParty(msg, params, keys[i], outCh, endCh).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	startGR := runtime.NumGoroutine()
	var ended int32
signing:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			var actualAttackers []int
			var expectedAttacker []int
			// as we blame the node in both bobMid and bobMidWc
			for _, el := range attackers {
				expectedAttacker = append(expectedAttacker, el)
				expectedAttacker = append(expectedAttacker, el)
			}

			if err.Victim().Index != attackers[0] && err.Victim().Index != attackers[1] {
				for _, el := range err.Culprits() {
					actualAttackers = append(actualAttackers, el.Index)
				}
				sort.Ints(expectedAttacker)
				sort.Ints(actualAttackers)
				assert.Equal(t, expectedAttacker, actualAttackers, "the victim should report all the attackers")
			}
			err.Culprits()
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Log("share share attack test finished successfully")
				t.Logf("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break signing
			}

		case msg := <-outCh:
			if msg.Type() == "SignRound1Message" && (msg.GetFrom().Index == attackers[0] || msg.GetFrom().Index == attackers[1]) {
				var attackContent SignRound1Message
				err = ptypes.UnmarshalAny(msg.WireMsg().Message, &attackContent)
				assert.Nil(t, err)
				localParty := parties[msg.GetFrom().Index]
				cA := localParty.temp.cAKI
				rA := localParty.temp.rAKI

				kIBz := parties[msg.GetFrom().Index].temp.r5AbortData.KI
				kI := new(big.Int).SetBytes(kIBz)

				paiPK := localParty.keys.PaillierPKs[msg.GetFrom().Index]
				incorrectID := msg.GetFrom().Index
				pi, err := mta.AliceInit(paiPK, kI, cA, rA, localParty.keys.NTildej[incorrectID], localParty.keys.H1j[incorrectID], localParty.keys.H2j[incorrectID])
				assert.Nil(t, err)

				localParty.temp.rangeProof = pi
				cmt := new(big.Int).SetBytes(attackContent.C)
				r1msg1 := NewSignRound1Message(localParty.PartyID(), cA, localParty.temp.rangeProof, cmt)
				localParty.temp.signRound1Messages[msg.GetFrom().Index] = r1msg1
				msg = r1msg1
			}
			for _, P := range parties {
				if P.PartyID().Index == msg.GetFrom().Index {
					continue
				}
				go updater(P, msg, errCh)
			}

		case <-endCh:
			t.Fatal("tss should never success")
		}
	}
}

func TestE2EConcurrentMtaAbortInRound5(t *testing.T) {
	setUp("info")
	attackers := []int{1, 2}
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endCh := make(chan *SignatureData, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	msg := common.GetRandomPrimeInt(256)
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		P := NewLocalParty(msg, params, keys[i], outCh, endCh).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	startGR := runtime.NumGoroutine()
	var ended int32
signing:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			var actualAttackers []int
			var expectedAttacker []int
			// as we blame the node in both bobMid and bobMidWc
			for _, el := range attackers {
				expectedAttacker = append(expectedAttacker, el)
			}

			if err.Victim().Index != attackers[0] && err.Victim().Index != attackers[1] {
				for _, el := range err.Culprits() {
					actualAttackers = append(actualAttackers, el.Index)
				}
				sort.Ints(expectedAttacker)
				sort.Ints(actualAttackers)
				assert.Equal(t, expectedAttacker, actualAttackers, "the victim should report all the attackers")
			}
			err.Culprits()
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Log("share share attack test finished successfully")
				t.Logf("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break signing
			}

		case msg := <-outCh:
			if msg.Type() == "SignRound2Message" && (msg.GetFrom().Index == attackers[0] || msg.GetFrom().Index == attackers[1]) {
				var attackContent SignRound2Message
				err = ptypes.UnmarshalAny(msg.WireMsg().Message, &attackContent)
				assert.Nil(t, err)
				localParty := parties[msg.GetFrom().Index]

				allItems := attackContent.Items
				allItems[0] = allItems[3]

				r2msg := NewSignRound2Message(localParty.PartyID(), allItems)
				localParty.temp.signRound2Messages[msg.GetFrom().Index] = r2msg
				msg = r2msg
			}
			for _, P := range parties {
				if P.PartyID().Index == msg.GetFrom().Index {
					continue
				}
				go updater(P, msg, errCh)
			}

		case <-endCh:
			t.Fatal("tss should never success")
		}
	}
}

func TestE2EConcurrentMtaAbortInRound52(t *testing.T) {
	setUp("info")
	attackers := []int{2}
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endCh := make(chan *SignatureData, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	msg := common.GetRandomPrimeInt(256)
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		P := NewLocalParty(msg, params, keys[i], outCh, endCh).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	startGR := runtime.NumGoroutine()
	var ended int32
signing:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			var actualAttackers []int
			var expectedAttacker []int
			// as we blame the node in both bobMid and bobMidWc
			for _, el := range attackers {
				expectedAttacker = append(expectedAttacker, el)
			}

			if err.Victim().Index != attackers[0] {
				for _, el := range err.Culprits() {
					actualAttackers = append(actualAttackers, el.Index)
				}
				sort.Ints(expectedAttacker)
				sort.Ints(actualAttackers)
				assert.Equal(t, expectedAttacker, actualAttackers, "the victim should report all the attackers")
			}
			err.Culprits()
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Log("share share attack test finished successfully")
				t.Logf("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break signing
			}

		case msg := <-outCh:
			if msg.Type() == "SignRound2Message" && msg.GetFrom().Index == attackers[0] {
				var attackContent SignRound2Message
				err = ptypes.UnmarshalAny(msg.WireMsg().Message, &attackContent)
				assert.Nil(t, err)
				localParty := parties[msg.GetFrom().Index]

				allItems := attackContent.Items
				allItems[0] = allItems[3]

				r2msg := NewSignRound2Message(localParty.PartyID(), allItems)
				localParty.temp.signRound2Messages[msg.GetFrom().Index] = r2msg
				msg = r2msg
			}
			for _, P := range parties {
				if P.PartyID().Index == msg.GetFrom().Index {
					continue
				}
				go updater(P, msg, errCh)
			}

		case <-endCh:
			t.Fatal("tss should never success")
		}
	}
}

func TestE2EConcurrentMtaAbortReceiverAttack(t *testing.T) {
	setUp("info")
	attackers := []int{2, 4}
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endCh := make(chan *SignatureData, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	msg := common.GetRandomPrimeInt(256)
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		P := NewLocalParty(msg, params, keys[i], outCh, endCh).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	startGR := runtime.NumGoroutine()
	var ended int32
signing:
	for {
		// fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			var actualAttackers []int
			var expectedAttacker []int
			// as we blame the node in both bobMid and bobMidWc
			for _, el := range attackers {
				expectedAttacker = append(expectedAttacker, el)
			}

			if err.Victim().Index != attackers[0] && err.Victim().Index != attackers[1] {
				for _, el := range err.Culprits() {
					actualAttackers = append(actualAttackers, el.Index)
				}
				sort.Ints(expectedAttacker)
				sort.Ints(actualAttackers)
				assert.Equal(t, expectedAttacker, actualAttackers, "the victim should report all the attackers")
			}
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Log("share share attack test finished successfully")
				t.Logf("Start goroutines: %d, End goroutines: %d", startGR, runtime.NumGoroutine())
				break signing
			}

		case msg := <-outCh:
			if msg.Type() == "SignRound3Message" && (msg.GetFrom().Index == attackers[0] || msg.GetFrom().Index == attackers[1]) {
				var attackContent SignRound3Message
				err = ptypes.UnmarshalAny(msg.WireMsg().Message, &attackContent)
				assert.Nil(t, err)
				localParty := parties[msg.GetFrom().Index]

				entry := SignRound3Message_AbortDataEntry{
					Index: int32(0),
				}

				abortMta := SignRound3Message_AbortData{
					Item: []*SignRound3Message_AbortDataEntry{&entry},
				}
				r3msg := NewSignRound3MessageAbort(localParty.PartyID(), &abortMta)
				localParty.temp.signRound3Messages[msg.GetFrom().Index] = r3msg
				msg = r3msg
			}
			for _, P := range parties {
				if P.PartyID().Index == msg.GetFrom().Index {
					continue
				}
				go updater(P, msg, errCh)
			}

		case <-endCh:
			t.Fatal("tss should never success")
		}
	}
}

func TestE2EConcurrentOneRound(t *testing.T) {
	setUp("info")
	threshold := testThreshold

	// PHASE: load keygen fixtures
	keys, signPIDs, err := keygen.LoadKeygenTestFixturesRandomSet(testThreshold+1, testParticipants)
	var ourOneRoundData *SignatureData_OneRoundData
	otherOneRoundData := make(map[*tss.PartyID]*SignatureData_OneRoundData)
	assert.NoError(t, err, "should load keygen fixtures")
	assert.Equal(t, testThreshold+1, len(keys))
	assert.Equal(t, testThreshold+1, len(signPIDs))

	// PHASE: signing
	// use a shuffled selection of the list of parties for this test
	p2pCtx := tss.NewPeerContext(signPIDs)
	parties := make([]*LocalParty, 0, len(signPIDs))

	errCh := make(chan *tss.Error, len(signPIDs))
	outCh := make(chan tss.Message, len(signPIDs))
	endChs := make([]chan *SignatureData, len(signPIDs))
	dataCh := make(chan signatureDataWithPartyId, len(signPIDs))

	updater := test.SharedPartyUpdater

	// init the parties
	for i := 0; i < len(signPIDs); i++ {
		params := tss.NewParameters(p2pCtx, signPIDs[i], len(signPIDs), threshold)

		endChs[i] = make(chan *SignatureData, 1)
		P := NewLocalParty(nil, params, keys[i], outCh, endChs[i]).(*LocalParty)
		parties = append(parties, P)
		go func(P *LocalParty) {
			if err := P.Start(); err != nil {
				errCh <- err
			}
		}(P)
	}

	var ended int32
	for i, endCh := range endChs {
		go func(ch chan *SignatureData, partyId *tss.PartyID) {
			for data := range ch {
				dataCh <- signatureDataWithPartyId{
					signData: data,
					partyId:  partyId,
				}
				break
			}
		}(endCh, signPIDs[i])
	}
preparing:
	for {
		fmt.Printf("ACTIVE GOROUTINES: %d\n", runtime.NumGoroutine())
		select {
		case err := <-errCh:
			common.Logger.Errorf("Error: %s", err)
			assert.FailNow(t, err.Error())
			break preparing

		case msg := <-outCh:
			dest := msg.GetTo()
			if dest == nil {
				for _, P := range parties {
					if P.PartyID().Index == msg.GetFrom().Index {
						continue
					}
					go updater(P, msg, errCh)
				}
			} else {
				if dest[0].Index == msg.GetFrom().Index {
					t.Fatalf("party %d tried to send a message to itself (%d)", dest[0].Index, msg.GetFrom().Index)
				}
				go updater(parties[dest[0].Index], msg, errCh)
			}

		case data := <-dataCh:
			for i, pid := range signPIDs {
				if pid.Id == data.partyId.Id {
					if i == 0 {
						ourOneRoundData = data.signData.OneRoundData
					} else {
						otherOneRoundData[signPIDs[i]] = data.signData.OneRoundData
					}
					break
				}
			}
			atomic.AddInt32(&ended, 1)
			if atomic.LoadInt32(&ended) == int32(len(signPIDs)) {
				t.Logf("Done. Received signature data from %d participants %+v", ended, data)
				break preparing
			}
		}
	}

	// Simulate offline round
	msg := common.GetRandomPrimeInt(256)
	otherSis := make(map[*tss.PartyID]*big.Int)
	ourSi := FinalizeGetOurSigShare(ourOneRoundData, msg)
	for partyId, data := range otherOneRoundData {
		start := time.Now()
		si := FinalizeGetOurSigShare(data, msg)
		t.Logf("compose si takes %d microseconds", time.Since(start).Microseconds())
		otherSis[partyId] = si
	}

	// compose final signature
	pkX, pkY := keys[0].ECDSAPub.X(), keys[0].ECDSAPub.Y()
	pk := ecdsa.PublicKey{
		Curve: tss.EC(),
		X:     pkX,
		Y:     pkY,
	}
	start := time.Now()
	_, sig, finalErr := FinalizeGetAndVerifyFinalSig(&SignatureData{OneRoundData: ourOneRoundData}, &pk, msg, signPIDs[0], ourSi, otherSis)
	t.Logf("calculate final sig takes %d microseconds", time.Since(start).Microseconds())
	assert.Nil(t, finalErr, "final signature generation should have no error")

	// BEGIN ECDSA verify
	ok := ecdsa.Verify(&pk, msg.Bytes(), sig.R, sig.S)
	assert.True(t, ok, "ecdsa verify must pass")

	btcecSig := &btcec.Signature{R: sig.R, S: sig.S}
	btcecSig.Verify(msg.Bytes(), (*btcec.PublicKey)(&pk))
	assert.True(t, ok, "ecdsa verify 2 must pass")

	t.Log("ECDSA signing test done.")
}
