Uniswap V2 里添加和移除流动性是一对相反的操作。添加是打进两种代币、按份额拿走 LP。移除是交还 LP、按同一比例拿回两种代币。

## 添加流动性

调用顺序：

1. 调用 Router 的 `addLiquidity`。
2. Router 先走 `_addLiquidity`。它问 Factory 要 `getPair`，交易对还不存在就 `createPair`。
3. Factory 用 `CREATE2` 部署 Pair。地址由两个代币地址算出来，所以部署前就能知道。之后 Router 用 `pairFor` 直接算出这个地址，不用再查一次。
4. Router 用两次 `transferFrom`，把两种代币从你的钱包转到 Pair。事先要授权给 Router。
5. Router 调用 `pair.mint(to)`。LP 在这一步直接铸给 `to`，没有另一次转账。

池子还是空的时，没有价格可跟，想打多少就打多少。第一次铸造会把 \(\sqrt{dx \cdot dy}\) 里的 1000 枚锁进零地址，剩下的才是你的 LP。这 1000 枚永久拿不回来，用来避免有人把份额抬得太便宜。

池子里已经有储备 \(x_0\)、\(y_0\)，LP 总量是 \(T\)。你打进 \(dx\)、\(dy\)。两种币必须按现有比例进：

$$
\frac{dx}{x_0} = \frac{dy}{y_0}
$$

拿到的 LP 取两边份额里较小的那个：

$$
S = \min\left(\frac{dx}{x_0},\ \frac{dy}{y_0}\right) \cdot T
$$

取最小值是为了不让你只多打一种币就多占份额。Router 会先用 `quote` 把数量收成这个比例，多出来的不会转进池子。`amountAMin` 和 `amountBMin` 是滑点下限，实际打进去的数量比这个还少，整笔交易回滚。

图上的 DAI/ETH 走的是 `addLiquidityETH`。ETH 不能直接进 Pair，Router 会先把它包成 WETH 再转进池子，多付的 ETH 退回。

## 移除流动性

调用顺序和添加相反：

1. 你调用 Router 的 `removeLiquidity`。Router 用 `pairFor` 算出池子地址，不再问 Factory。
2. Router 用 `transferFrom` 把你的 LP 转到 Pair 自己手里。所以也要先把 LP 授权给 Router。`burn` 只认已经躺在 Pair 合约里的 LP，不会直接从你钱包里烧。
3. Router 调用 `pair.burn(to)`。
4. `burn` 按份额把两种代币打给 `to`，再把刚收到的那笔 LP 烧掉。

设烧掉的 LP 是 \(S\)，总量是 \(T\)，池子储备是 \(x_0\)、\(y_0\)：

$$
dx = \frac{S}{T}\, x_0, \quad dy = \frac{S}{T}\, y_0
$$

这就是代码里的 `liquidity * balance / totalSupply`。这里用的是池子里两种代币的余额，不是储备。有人直接给池子转过额外的币，也会按同一比例分掉。

两种币按同一个比例退出，价格不变：

$$
\frac{x_0}{y_0} = \frac{x_0 - dx}{y_0 - dy}
$$

流动性深度 \(L = \sqrt{x \cdot y}\) 也按同一比例缩小。\(L_1 = \left(1 - \dfrac{S}{T}\right) L_0\)，所以：

$$
\frac{L_0 - L_1}{L_0} = \frac{S}{T}
$$

退完后 Router 再核对 `amountAMin` 和 `amountBMin`。价格已经变了、退回来太少，整笔交易回滚。

图上的 DAI/ETH 走的是 `removeLiquidityETH`。它先把两种币 `burn` 到 Router，再把 WETH 拆成 ETH，连同另一种代币一起转给你。普通的 `removeLiquidity` 不拆，两种 ERC20 直接打给 `to`。