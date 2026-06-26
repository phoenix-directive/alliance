package v6

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/terra-money/alliance/x/alliance/types"
)

func Migrate(sk types.StakingKeeper, bk types.BankKeeper) func(ctx sdk.Context) error {
	return func(ctx sdk.Context) error {
		bondedValidatorTokens, err := bondedValidatorTokens(ctx, sk)
		if err != nil {
			return err
		}

		bondedPoolTokens, err := sk.TotalBondedTokens(ctx)
		if err != nil {
			return err
		}

		if bondedValidatorTokens.Equal(bondedPoolTokens) {
			return nil
		}

		bondDenom, err := sk.BondDenom(ctx)
		if err != nil {
			return err
		}

		if bondedValidatorTokens.GT(bondedPoolTokens) {
			diff := bondedValidatorTokens.Sub(bondedPoolTokens)
			return bk.SendCoinsFromModuleToModule(
				ctx,
				stakingtypes.NotBondedPoolName,
				stakingtypes.BondedPoolName,
				sdk.NewCoins(sdk.NewCoin(bondDenom, diff)),
			)
		}

		diff := bondedPoolTokens.Sub(bondedValidatorTokens)
		return bk.SendCoinsFromModuleToModule(
			ctx,
			stakingtypes.BondedPoolName,
			stakingtypes.NotBondedPoolName,
			sdk.NewCoins(sdk.NewCoin(bondDenom, diff)),
		)
	}
}

func bondedValidatorTokens(ctx sdk.Context, sk types.StakingKeeper) (math.Int, error) {
	validators, err := sk.GetAllValidators(ctx)
	if err != nil {
		return math.Int{}, err
	}

	total := math.ZeroInt()
	for _, validator := range validators {
		if validator.IsBonded() {
			total = total.Add(validator.Tokens)
		}
	}

	return total, nil
}
