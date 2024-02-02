package connection_test

import (
	"context"
	"encoding/hex"
	"fmt"
	"math/big"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	// "github.com/ethereum/go-ethereum/common"
	"github.com/avast/retry-go/v4"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"
	node_deposit "github.com/stafiprotocol/eth-lsd-relay/bindings/NodeDeposit"
	"github.com/stafiprotocol/eth-lsd-relay/pkg/connection"
	"github.com/stafiprotocol/eth-lsd-relay/pkg/connection/beacon"
	"github.com/stafiprotocol/eth-lsd-relay/pkg/connection/beacon/client"
	"github.com/stafiprotocol/eth-lsd-relay/pkg/connection/types"
	"github.com/stafiprotocol/eth-lsd-relay/pkg/utils"
)

func TestCallOpts(t *testing.T) {
	c, err := connection.NewConnection(os.Getenv("ETH1_ENDPOINT"), os.Getenv("ETH2_ENDPOINT"), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	oldopts := c.CallOpts(nil)
	t.Log(oldopts)
	newopts := c.CallOpts(big.NewInt(5))
	t.Log(oldopts)
	t.Log(newopts)

	newopts2 := c.CallOpts(big.NewInt(7))
	t.Log(oldopts)
	t.Log(newopts)
	t.Log(newopts2)

	gasPrice, err := c.Eth1Client().SuggestGasPrice(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	gasTip, err := c.Eth1Client().SuggestGasTipCap(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(gasPrice.String(), gasTip.String())

	beaconBlock, exist, err := c.GetBeaconBlock(5145404)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(beaconBlock.FeeRecipient, exist)

}

func TestBlockReward(t *testing.T) {
	c, err := connection.NewConnection(os.Getenv("ETH1_ENDPOINT"), os.Getenv("ETH2_ENDPOINT"), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	eth1Block, err := c.Eth1Client().BlockByNumber(context.Background(), big.NewInt(859542))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", eth1Block.Coinbase())
	totalFee := big.NewInt(0)
	for _, tx := range eth1Block.Transactions() {
		receipt, err := c.Eth1Client().TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			t.Fatal(err)
		}

		priorityGasFee := tx.EffectiveGasTipValue(eth1Block.BaseFee())

		totalFee = new(big.Int).Add(totalFee, new(big.Int).Mul(priorityGasFee, big.NewInt(int64(receipt.GasUsed))))
	}
	t.Log(totalFee)

	eth1Block, err = c.Eth1Client().BlockByNumber(context.Background(), big.NewInt(859543))
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("%v", eth1Block.Coinbase())
	totalFee = big.NewInt(0)
	for _, tx := range eth1Block.Transactions() {
		receipt, err := c.Eth1Client().TransactionReceipt(context.Background(), tx.Hash())
		if err != nil {
			t.Fatal(err)
		}

		priorityGasFee := tx.EffectiveGasTipValue(eth1Block.BaseFee())

		totalFee = new(big.Int).Add(totalFee, new(big.Int).Mul(priorityGasFee, big.NewInt(int64(receipt.GasUsed))))
	}
	t.Log(totalFee)

}

func TestBlockDetail(t *testing.T) {
	s := make([]int64, 0)
	sort.SliceStable(s, func(i, j int) bool { return s[i] < s[j] })

	logrus.SetLevel(logrus.DebugLevel)
	c, err := connection.NewConnection(os.Getenv("ETH1_ENDPOINT"), os.Getenv("ETH2_ENDPOINT"), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	beaconBlock, _, err := c.GetBeaconBlock(7312423)
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("%+v", beaconBlock)
	for _, tx := range beaconBlock.Transactions {
		t.Logf("%s", hex.EncodeToString(tx.TxHash))
		t.Logf("%s", hex.EncodeToString(tx.Recipient))
		t.Logf("%s", new(big.Int).SetBytes(tx.Amount).String())
	}

	return

	r, err := c.Eth2Client().SyncCommitteeRewards(6190497)
	if err != nil {
		if err != nil {
			switch {
			case strings.Contains(err.Error(), client.ErrBlockNotFound.Error()):
				// block not exit, should return
				t.Log("not exit")
			case strings.Contains(err.Error(), client.ErrSlotPreSyncCommittees.Error()):
				// skip err
				t.Log("skip")
			default:
				t.Log(err)
			}
		}
		t.Fatal(err)
	}
	t.Log(r)
	return

	balance, err := c.Eth2Client().Balance(77999, 61730)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(balance)
	return

	head, err := c.Eth2BeaconHead()
	if err != nil {
		t.Fatal(err)
	}

	pubkey, _ := types.HexToValidatorPubkey("93ce5068db907b2e5055dbb7805a3a3d7c56c9e82d010e864403e10a61235db4795949f01302dc2ad2b6225963599ed5")
	status, err := c.Eth2Client().GetValidatorStatus(pubkey, &beacon.ValidatorStatusOptions{
		Epoch: new(uint64),
		Slot:  &head.Slot,
	})
	if err != nil {
		t.Fatal(err)
	}

	t.Log(hex.EncodeToString(status.WithdrawalCredentials.Bytes()))
	eth1Block, err := c.Eth1Client().BlockByNumber(context.Background(), big.NewInt(190767))
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range eth1Block.Withdrawals() {
		t.Logf("%+v", w)

	}

	beaconBlock, _, err = c.Eth2Client().GetBeaconBlock(199214)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%+v", beaconBlock.Withdrawals)
	config, err := c.Eth2Client().GetEth2Config()
	timestamp := utils.StartTimestampOfEpoch(config, 10383)
	t.Log(timestamp)

}

func TestBalance(t *testing.T) {
	cc, err := ethclient.Dial(os.Getenv("ETH1_ENDPOINT"))
	if err != nil {
		t.Fatal(err)
	}
	blockNumber, err := cc.BlockNumber(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(blockNumber)
	tx, err := cc.TransactionReceipt(context.Background(), common.HexToHash("0x7e1bd5879335a0bc5d088f7709d76ba257de6b00473bc441c65fa9eedd552e57"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(tx.Logs)

	return
	c, err := connection.NewConnection(os.Getenv("ETH1_ENDPOINT"), os.Getenv("ETH2_ENDPOINT"), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	startSlot := uint64(204864)
	endSlot := uint64(204895)
	withdrawSlot := uint64(204886)
	epoch := uint64(6402)

	startStatus, err := c.GetValidatorStatusByIndex(fmt.Sprint(62947), &beacon.ValidatorStatusOptions{
		Slot: &startSlot,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(startStatus.Balance)

	withdrawStatus, err := c.GetValidatorStatusByIndex(fmt.Sprint(62947), &beacon.ValidatorStatusOptions{
		Slot: &withdrawSlot,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(withdrawStatus.Balance)

	endStatus, err := c.GetValidatorStatusByIndex(fmt.Sprint(62947), &beacon.ValidatorStatusOptions{
		Slot: &endSlot,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(endStatus.Balance)

	epochStatus, err := c.GetValidatorStatusByIndex(fmt.Sprint(62947), &beacon.ValidatorStatusOptions{
		Epoch: &epoch,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Log(epochStatus.Balance)
	t.Log(epochStatus.Status)
	t.Logf("%+v", epochStatus)

	config, err := c.Eth2Client().GetEth2Config()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(utils.StartSlotOfEpoch(config, epoch))
}

func TestGettingFirstNodeStakeEvent(t *testing.T) {
	c, err := connection.NewConnection(os.Getenv("ETH1_ENDPOINT"), os.Getenv("ETH2_ENDPOINT"), nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var start = uint64(0)
	latestBlock, err := c.Eth1LatestBlock()
	if err != nil {
		t.Fatal(err)
	}
	end := latestBlock

	nodeDeposits := []string{
		"0x179386303fC2B51c306Ae9D961C73Ea9a9EA0C8d",
		"0x8A57bC7fB1237f9fBF075A261Ed28F04105Cd89d",
	}

	for _, nodeDepositAddr := range nodeDeposits {
		fmt.Println("nodeDepositAddr:", nodeDepositAddr)

		nodeDepositContract, err := node_deposit.NewNodeDeposit(common.HexToAddress(nodeDepositAddr), c.Eth1Client())
		if err != nil {
			t.Fatal(err)
		}
		iter, err := retry.DoWithData(func() (*node_deposit.NodeDepositStakedIterator, error) {
			return nodeDepositContract.FilterStaked(&bind.FilterOpts{
				Start:   start,
				End:     &end,
				Context: context.Background(),
			})
		}, retry.Delay(time.Second*2), retry.Attempts(150))
		if err != nil {
			t.Fatal(err)
		}

		// for iter.Next() {
		// 	fmt.Println("stake event at:", iter.Event.Raw.BlockNumber)
		// }

		hasEvent := iter.Next()
		iter.Close()
		if hasEvent {
			// found the first node deposit event
			fmt.Println("first stake event", iter.Event.Raw.BlockNumber)
		} else {
			fmt.Println("no node stake event")
		}
	}
	// lsdTokens: 0x37a7BF277f9b1F32296aB595600eA30c55F6eE4B
	// lsdTokens: 0xD2a1e6931e8a41043cE80C4F7EB0F7083E64Bfb8 ( created by robert)
}
