package keeper

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/petrinetwork/petrichor/x/petrichor/types"
)

func (k Keeper) GetPetrichorValidator(ctx sdk.Context, valAddr sdk.ValAddress) (types.PetrichorValidator, error) {
	val, found := k.stakingKeeper.GetValidator(ctx, valAddr)
	if !found {
		return types.PetrichorValidator{}, fmt.Errorf("validator with address %s does not exist", valAddr.String())
	}
	valInfo, found := k.GetPetrichorValidatorInfo(ctx, valAddr)
	if !found {
		valInfo = k.createPetrichorValidatorInfo(ctx, valAddr)
	}
	return types.PetrichorValidator{
		Validator:             &val,
		PetrichorValidatorInfo: &valInfo,
	}, nil
}

func (k Keeper) GetPetrichorValidatorInfo(ctx sdk.Context, valAddr sdk.ValAddress) (types.PetrichorValidatorInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetrichorValidatorInfoKey(valAddr)
	vb := store.Get(key)
	var info types.PetrichorValidatorInfo
	if vb == nil {
		return info, false
	} else {
		k.cdc.MustUnmarshal(vb, &info)
		return info, true
	}
}

func (k Keeper) createPetrichorValidatorInfo(ctx sdk.Context, valAddr sdk.ValAddress) (val types.PetrichorValidatorInfo) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetPetrichorValidatorInfoKey(valAddr)
	val = types.NewPetrichorValidatorInfo()
	vb := k.cdc.MustMarshal(&val)
	store.Set(key, vb)
	return val
}

func (k Keeper) IteratePetrichorValidatorInfo(ctx sdk.Context, cb func(valAddr sdk.ValAddress, info types.PetrichorValidatorInfo) (stop bool)) {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorInfoKey)
	defer iter.Close()
	for ; iter.Valid(); iter.Next() {
		var info types.PetrichorValidatorInfo
		b := iter.Value()
		k.cdc.MustUnmarshal(b, &info)
		valAddr := types.ParsePetrichorValidatorKey(iter.Key())
		if cb(valAddr, info) {
			return
		}
	}
}

func (k Keeper) GetAllPetrichorValidatorInfo(ctx sdk.Context) []types.PetrichorValidatorInfo {
	store := ctx.KVStore(k.storeKey)
	iter := sdk.KVStorePrefixIterator(store, types.ValidatorInfoKey)
	defer iter.Close()
	var infos []types.PetrichorValidatorInfo
	for ; iter.Valid(); iter.Next() {
		b := iter.Value()
		var info types.PetrichorValidatorInfo
		k.cdc.UnmarshalInterface(b, &info)
		infos = append(infos, info)
	}
	return infos
}
