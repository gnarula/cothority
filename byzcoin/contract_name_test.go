package byzcoin

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.dedis.ch/cothority/v3"
	"go.dedis.ch/cothority/v3/darc"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
)

// TestService_Naming tries to name the genesis darc instance with many failed
// attempts before eventually passing.
func TestService_Naming(t *testing.T) {
	local := onet.NewTCPTest(cothority.Suite)
	defer local.CloseAll()

	signer := darc.NewSignerEd25519(nil, nil)
	hosts, roster, _ := local.GenTree(4, true)
	s := local.GetServices(hosts, ByzCoinID)[0].(*Service)

	genesisMsg, err := DefaultGenesisMsg(CurrentVersion, roster, []string{"_name:" + ContractDarcID}, signer.Identity())
	require.Nil(t, err)
	gDarc := &genesisMsg.GenesisDarc
	genesisMsg.BlockInterval = time.Second
	cl, _, err := NewLedger(genesisMsg, false)
	require.Nil(t, err)

	var namingTx ClientTransaction

	// FAIL - use a bad signature
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{1},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	namingTx.Instructions[0].Signatures[0] = append(namingTx.Instructions[0].Signatures[0][1:], 0) // tamper the signature
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.Error(t, err)

	// FAIL - use a use an instance that does not exist
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: append(gDarc.GetBaseID()[1:], 0), // does not exist
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{1},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.Error(t, err)

	// FAIL - use a signer that is not authorized by the instance to spawn
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{1},
			},
		},
	}
	signer2 := darc.NewSignerEd25519(nil, nil) // bad signer
	require.NoError(t, namingTx.FillSignersAndSignWith(signer2))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.Error(t, err)

	// FAIL - missing instance name
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte{}, // missing name
						},
					},
				},
				SignerCounter: []uint64{1},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.Error(t, err)

	// SUCCEED - Make one name and it should succeed.
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{1},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.NoError(t, err)

	// FAIL - Overwriting the name is not allowed (it must be deleted first).
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{2},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.Error(t, err)

	// SUCCEED - Making multiple names is allowed.
	namingTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("your genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{2},
			},
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "add",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("everyone's genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{3},
			},
		},
	}
	require.NoError(t, namingTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(namingTx, 10)
	require.NoError(t, err)

	// Check that the names for a chain.
	rst, err := s.GetReadOnlyStateTrie(cl.ID)
	require.NoError(t, err)
	buf, _, cID, dID, err := rst.GetValues(NamingInstanceID[:])
	require.NoError(t, err)
	require.Equal(t, "naming", cID)
	require.Empty(t, dID)

	name := ContractNamingBody{}
	require.NoError(t, protobuf.Decode(buf, &name))
	latest := name.Latest
	require.NotEqual(t, latest, InstanceID{})

	var cnt int
	for !latest.Equal(InstanceID{}) {
		buf, _, cID, dID, err = rst.GetValues(latest[:])
		require.NoError(t, err)
		require.Empty(t, cID)
		require.Empty(t, dID)

		entry := contractNamingEntry{}
		require.NoError(t, protobuf.Decode(buf, &entry))

		latest = entry.Prev
		cnt++
	}
	// Count should be 3 because we sent 3 good instructions.
	require.Equal(t, 3, cnt)

	// Use the API to get those names back
	// Wrong name
	_, err = cl.ResolveInstanceID(gDarc.GetBaseID(), "wrong name")
	require.Error(t, err)

	// Wrong darc ID
	_, err = cl.ResolveInstanceID(append(gDarc.GetBaseID()[1:], 0), "my genesis darc")
	require.Error(t, err)

	// Correct
	verifyNameResolution := func(name string) {
		var iID InstanceID
		var proofResp *GetProofResponse
		iID, err = cl.ResolveInstanceID(gDarc.GetBaseID(), name)
		require.NoError(t, err)
		proofResp, err = cl.GetProof(iID[:])
		require.NoError(t, err)
		require.NoError(t, proofResp.Proof.Verify(cl.ID))
	}
	verifyNameResolution("my genesis darc")
	verifyNameResolution("your genesis darc")
	verifyNameResolution("everyone's genesis darc")

	// Tests below are for removal.

	// FAIL - do not allow removal for what does not exist.
	removalTx := ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "remove",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("not exists"),
						},
					},
				},
				SignerCounter: []uint64{4},
			},
		},
	}
	require.NoError(t, removalTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(removalTx, 10)
	require.Error(t, err)

	// SUCCESS - try to remove an entry.
	removalTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "remove",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{4},
			},
		},
	}
	require.NoError(t, removalTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(removalTx, 10)
	require.NoError(t, err)

	// FAIL - the removed entry cannot be "removed" again.
	removalTx = ClientTransaction{
		Instructions: Instructions{
			{
				InstanceID: NamingInstanceID,
				Invoke: &Invoke{
					ContractID: ContractNamingID,
					Command:    "remove",
					Args: Arguments{
						{
							Name:  "instanceID",
							Value: gDarc.GetBaseID(),
						},
						{
							Name:  "name",
							Value: []byte("my genesis darc"),
						},
					},
				},
				SignerCounter: []uint64{5},
			},
		},
	}
	require.NoError(t, removalTx.FillSignersAndSignWith(signer))
	_, err = cl.AddTransactionAndWait(removalTx, 10)
	require.Error(t, err)

	// Try to resolve the deleted entry should fail, but the others should
	// exist.
	_, err = cl.ResolveInstanceID(gDarc.GetBaseID(), "my genesis darc")
	require.Error(t, err)

	verifyNameResolution("your genesis darc")
	verifyNameResolution("everyone's genesis darc")
}
