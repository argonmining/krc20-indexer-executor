
////////////////////////////////
package operation

import (
    "strings"
    "strconv"
    "math/big"
    "kasplex-executor/misc"
    "kasplex-executor/storage"
)

////////////////////////////////
type OpMethodList struct {}

////////////////////////////////
func init() {
    opName := "list"
    P_Registered["KRC-20"] = true
    Op_Registered[opName] = true
    Method_Registered[opName] = new(OpMethodList)
}

////////////////////////////////
func (opMethodList OpMethodList) FeeLeast(daaScore uint64) (uint64) {
    return 0
}

////////////////////////////////
func (opMethodList OpMethodList) ScriptCollectEx(index int, script *storage.DataScriptType, txData *storage.DataTransactionType, testnet bool) {
    script.Utxo = ""
    if len(txData.Data.Outputs) > 0 {
        script.Utxo = txData.TxId + "_" + txData.Data.Outputs[0].VerboseData.ScriptPublicKeyAddress + "_" + strconv.FormatUint(txData.Data.Outputs[0].Amount,10)
    }
}

////////////////////////////////
func (opMethodList OpMethodList) Validate(script *storage.DataScriptType, txId string, daaScore uint64, testnet bool) (bool) {
    if (!testnet && daaScore < 97539090) {
        return false
    }
    if ValidateTxId(&script.Sc) {
        script.Tick = script.Sc
    }
    if (script.From == "" || script.Utxo == "" || script.P != "KRC-20" || !ValidateTickTxId(&script.Tick) || !ValidateAmount(&script.Amt)) {
        return false
    }
    script.To = ""
    script.Max = ""
    script.Lim = ""
    script.Pre = ""
    script.Dec = ""
    script.Price = ""
    script.Mod = ""
    script.Name = ""
    script.Sc = ""
    return true
}

////////////////////////////////
func (opMethodList OpMethodList) PrepareStateKey(opScript *storage.DataScriptType, stateMap storage.DataStateMapType) {
    stateMap.StateTokenMap[opScript.Tick] = nil
    stateMap.StateBalanceMap[opScript.From+"_"+opScript.Tick] = nil
    stateMap.StateBlacklistMap[opScript.Tick+"_"+opScript.From] = nil
}

////////////////////////////////
func (opMethodList OpMethodList) Do(index int, opData *storage.DataOperationType, stateMap storage.DataStateMapType, testnet bool) (error) {
    opScript := opData.OpScript[index]
    ////////////////////////////////
    if stateMap.StateTokenMap[opScript.Tick] == nil {
        opData.OpAccept = -1
        opData.OpError = "tick not found"
        return nil
    }
    if stateMap.StateBlacklistMap[opScript.Tick+"_"+opScript.From] != nil {
        opData.OpAccept = -1
        opData.OpError = "blacklist"
        return nil
    }
    ////////////////////////////////
    dataUtxo := strings.Split(opScript.Utxo, "_")
    keyMarket := opScript.Tick +"_"+ opScript.From +"_"+ dataUtxo[0]
    keyBalance := opScript.From +"_"+ opScript.Tick
    stBalance := stateMap.StateBalanceMap[keyBalance]
    opScript.Name = stateMap.StateTokenMap[opScript.Tick].Name
    ////////////////////////////////
    if stBalance == nil {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    }
    balanceBig := new(big.Int)
    balanceBig.SetString(stBalance.Balance, 10)
    amtBig := new(big.Int)
    amtBig.SetString(opScript.Amt, 10)
    if amtBig.Cmp(balanceBig) > 0 {
        opData.OpAccept = -1
        opData.OpError = "balance insuff"
        return nil
    }
    uJson := ""
    if stateMap.StateTokenMap[opScript.Tick].Mod == "issue" {
        uJson = `{"p":"krc-20","op":"send","sc":"` + strings.ToLower(opScript.Tick) + `"}`
    } else {
        uJson = `{"p":"krc-20","op":"send","tick":"` + strings.ToLower(opScript.Tick) + `"}`
    }
    uAddr, uScript := misc.MakeP2shKasplex(opData.ScriptSig, "", uJson, testnet)
    if dataUtxo[1] != uAddr {
        opData.OpAccept = -1
        opData.OpError = "address invalid"
        return nil
    }
    ////////////////////////////////
    opData.StBefore = nil
    opData.StBefore = AppendStLineBalance(opData.StBefore, keyBalance, stBalance, false)
    opData.StBefore = AppendStLineMarket(opData.StBefore, keyMarket, nil, false)
    ////////////////////////////////
    balanceBig = balanceBig.Sub(balanceBig, amtBig)
    lockedBig := new(big.Int)
    lockedBig.SetString(stBalance.Locked, 10)
    lockedBig = lockedBig.Add(lockedBig, amtBig)
    stBalance.Balance = balanceBig.Text(10)
    stBalance.Locked = lockedBig.Text(10)
    stBalance.OpMod = opData.OpScore
    stMarket := &storage.StateMarketType{
        Tick: opScript.Tick,
        TAmt: opScript.Amt,
        TAddr: opScript.From,
        UTxId: dataUtxo[0],
        UAddr: dataUtxo[1],
        UAmt: dataUtxo[2],
        UScript: uScript,
        OpAdd: opData.OpScore,
    }
    stateMap.StateMarketMap[keyMarket] = stMarket
    ////////////////////////////////
    opData.StAfter = nil
    opData.StAfter = AppendStLineBalance(opData.StAfter, keyBalance, stBalance, true)
    opData.StAfter = AppendStLineMarket(opData.StAfter, keyMarket, stMarket, true)
    ////////////////////////////////
    opData.OpAccept = 1
    return nil
}

////////////////////////////////
/*func (opMethodList OpMethodList) UnDo() (error) {
    // ...
    return nil
}*/

// ...
