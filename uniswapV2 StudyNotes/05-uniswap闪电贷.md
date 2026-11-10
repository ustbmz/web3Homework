# Uniswap V2 Flash Swap

Flash swap 是同一笔交易里先把币借走，回调里做完事情再还上。还不上，或者还完之后不满足恒定乘积，整笔交易回滚，池子回到调用前的状态。不需要抵押，因为借和还发生在同一笔交易里。

普通 swap 是先把输入代币打进 Pair，再调用 `swap`，而且 `data` 为空。Flash swap 也是调用同一个 `swap`，区别只在 `data` 不为空：Pair 会先把输出代币转走，再回调收款合约。

## 学习路线

1. 分清两种用法。`data` 为空是普通兑换，调用前输入代币已经在 Pair 里。`data` 不为空才是 flash swap。
2. 顺着 `UniswapV2Pair.swap` 读四步：转出、回调、计量、检查恒定乘积。
3. 算清「借一种币、还同一种币」要多还多少。这里用的是 1000/997，不是普通兑换里的 997/1000。
4. 看回调接口 `IUniswapV2Callee.uniswapV2Call`。回调里必须确认 `msg.sender` 真的是那个 Pair。
5. 读仓库里的例子 `contracts/examples/ExampleFlashSwap.sol`。它从 V2 借币，到 V1 换，再把欠款还回 V2，差额留下。
6. 最后再看「把手续费从 0.3% 压到 0.05%」那个例子。那是 flash swap 的一种用法，不是手续费规则本身改了。

## 调用顺序

图上四步就是 `swap` 里的顺序。

1. 你的合约调用 `pair.swap(amount0Out, amount1Out, to, data)`。`to` 必须是实现了 `uniswapV2Call` 的合约，不能是两个代币合约自己。`data` 传空字节以外的内容，才会触发回调。
2. Pair 先把 `amount0Out`、`amount1Out` 转给 `to`。这时候钱已经离开池子，检查还没做。
3. Pair 回调 `to.uniswapV2Call(msg.sender, amount0Out, amount1Out, data)`。第一个参数是发起 `swap` 的那个合约，不是 Pair。
4. 回调返回后，Pair 看自己余额比「储备减去借出」多出来多少，多出来的就是还回来的输入。然后做 \(k\) 检查。不过就回滚。

```170:182:contracts/UniswapV2Pair.sol
        if (amount0Out > 0) _safeTransfer(_token0, to, amount0Out); // optimistically transfer tokens
        if (amount1Out > 0) _safeTransfer(_token1, to, amount1Out); // optimistically transfer tokens
        if (data.length > 0) IUniswapV2Callee(to).uniswapV2Call(msg.sender, amount0Out, amount1Out, data);
        balance0 = IERC20(_token0).balanceOf(address(this));
        balance1 = IERC20(_token1).balanceOf(address(this));
        ...
        uint amount0In = balance0 > _reserve0 - amount0Out ? balance0 - (_reserve0 - amount0Out) : 0;
        uint amount1In = balance1 > _reserve1 - amount1Out ? balance1 - (_reserve1 - amount1Out) : 0;
        require(amount0In > 0 || amount1In > 0, 'UniswapV2: INSUFFICIENT_INPUT_AMOUNT');
```

还哪种币都可以。借 token0、还 token1，就是一笔普通兑换，只是付款发生在收款之后。借 token0、仍还 token0，就是无抵押借币，多还的那一截是手续费。

## 借同一种币时要还多少

设池子里这种币原来是 \(x_0\)，借走 \(dx_0\)。回调里还回来 \(dx\)。交易结束时池子余额是：

$$
x_0 - dx_0 + dx
$$

手续费按还回来的数量收 0.30%。要让 \(k\) 不下降，还回来的数量必须满足：

$$
dx \ge dx_0 \cdot \frac{1000}{997}
$$

多出来的就是手续费：

$$
\text{fee} = dx - dx_0 = dx_0 \cdot \frac{3}{997}
$$

3/997 约是 0.301%。图上写的 0.003 倍是近似，精确倍数是 3/997。

不要和普通兑换的 997 搞反。普通兑换是你先付 \(dx\)，只有 \(0.997\,dx\) 参与计价。Flash 借同一种币是你已经拿走了 \(dx_0\)，所以要倒过来乘 \(\dfrac{1000}{997}\)。

代码里的调整就是在做这件事。余额乘 1000，再扣掉输入的 0.3%：

$$
\text{balanceAdjusted} = \text{balance} \times 1000 - \text{amountIn} \times 3
$$

两种币的调整后余额相乘，必须不小于原来的 \(k \times 1000^2\)。不过就报 `UniswapV2: K`。

## 回调里要做的事

`uniswapV2Call` 跑在借到币之后、检查 \(k\) 之前。最少三件事：

1. 用 `pairFor` 核对 `msg.sender` 是不是工厂名下的那个 Pair。谁都能直接调用这个函数，不核对的话，假的 Pair 可以骗你的合约转账。
2. 用借到的币做这笔交易里要做的事。
3. 把欠的数量转回 `msg.sender`，也就是 Pair。转少了，后面的 \(k\) 检查会失败。

仓库里的 `ExampleFlashSwap` 就是这个结构。它只做单向、且必须有一边是 WETH：从 V2 借出，到对应的 V1 交易所换成另一种，用 `getAmountsIn` 算出要还 V2 的数量，多出来的留给发起人。还不够就 `assert` 失败，整笔回滚。

## 把 0.30% 做到接近 0.05%

这是 flash swap 的一个用法，不是把池子费率改掉。

0.30% 里有 5/6 归流动性提供者，1/6 归协议。1/6 乘 0.30% 等于 0.05%，5/6 约是 83%。

如果在同一笔交易里先变成这个池子几乎全部的流动性提供者，再去做那笔兑换，自己付出去的手续费里大约 83% 会按 LP 份额回到自己手里。净成本就接近协议留下的那 0.05%。做完再撤出流动性，还掉最初借的币。

图上那五步就是这个想法：借来一大笔资金，拿去铸造 DAI 并灌进 DAI 池，占住绝大部分份额，再做目标交易，最后归还借款。池子里本来已经有别人的流动性时，你分到的比例就没有 83% 那么高。
