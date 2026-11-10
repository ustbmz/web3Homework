Uniswap V2 的手续费分两层。交易时收 0.30%，留在池子里，所以 \(k = x \cdot y\) 变大，流动性提供者一起受益。工厂如果设了 `feeTo`，协议再从这笔增长里拿走 \(\dfrac{1}{6}\)，也就是全部交易额的 0.05%，剩下 \(\dfrac{5}{6}\) 仍归流动性提供者。

## 交易费怎么让 \(k\) 变大

用 DAI 换 ETH 时，先从打进池子的 DAI 里扣 0.30%，只用剩下的 99.7% 去推 \(x \cdot y = k\)。Router 里就是乘 997、除 1000：

```43:49:contracts/libraries/UniswapV2Library.sol
    function getAmountOut(uint amountIn, uint reserveIn, uint reserveOut) internal pure returns (uint amountOut) {
        ...
        uint amountInWithFee = amountIn.mul(997);
        uint numerator = amountInWithFee.mul(reserveOut);
        uint denominator = reserveIn.mul(1000).add(amountInWithFee);
        amountOut = numerator / denominator;
    }
```

扣下的 0.30% 没有转走，还在池子里。所以换完之后，真实余额的乘积大于换之前的 \(k\)。

## 流动性提供者怎么受益

记换之前 \(k_1\)，换完并同步储备之后 \(k_2\)。池子深度 \(L = \sqrt{k}\) 的相对增长是：

$$
f_{1,2} = 1 - \sqrt{\frac{k_1}{k_2}} = \frac{\sqrt{k_2} - \sqrt{k_1}}{\sqrt{k_2}}
$$

你的 LP 数量没变，但每个 LP 对应的池子变厚了。撤出时按份额拿，两种币都会比「没有手续费」时多一点。这就是图上「让 LP 受益于手续费」的那一块。

## 协议拿走 \(\phi = \dfrac{1}{6}\)

\(\phi\) 是项目方想从这笔手续费增长里分走的比例，V2 固定为 \(\dfrac{1}{6}\)。

时间从 \(t_1\) 到 \(t_2\)，\(\sqrt{k}\) 从 \(\sqrt{k_1}\) 涨到 \(\sqrt{k_2}\)，当时已有 LP 总量是 \(S_1\)。要铸给 `feeTo` 的 LP 数量是：

$$
S_m = \frac{\sqrt{k_2} - \sqrt{k_1}}{\left(\dfrac{1}{\phi} - 1\right)\sqrt{k_2} + \sqrt{k_1}} \cdot S_1
$$

把 \(\phi = \dfrac{1}{6}\) 代进去，\(\dfrac{1}{\phi} - 1 = 5\)：

$$
S_m = \frac{\sqrt{k_2} - \sqrt{k_1}}{5\sqrt{k_2} + \sqrt{k_1}} \cdot S_1
$$

分母里的 5 就是「流动性提供者留 5 份，协议拿 1 份」。

## 对应代码

`mint` 和 `burn` 一开始都会先调用 `_mintFee`。它比的是现在的 \(\sqrt{\text{reserve}_0 \cdot \text{reserve}_1}\) 和上次流动性变动时记下的 \(\sqrt{k_{\text{Last}}}\)。

```94:101:contracts/UniswapV2Pair.sol
            if (_kLast != 0) {
                uint rootK = Math.sqrt(uint(_reserve0).mul(_reserve1));
                uint rootKLast = Math.sqrt(_kLast);
                if (rootK > rootKLast) {
                    uint numerator = totalSupply.mul(rootK.sub(rootKLast));
                    uint denominator = rootK.mul(5).add(rootKLast);
                    uint liquidity = numerator / denominator;
```

分子是 \(S_1(\sqrt{k_2} - \sqrt{k_1})\)，分母是 \(5\sqrt{k_2} + \sqrt{k_1}\)，和上面的 \(S_m\) 是同一个式子。铸完之后，这次 `mint` 或 `burn` 结束时把新的 \(k\) 写进 `kLast`。`feeTo` 是零地址时不铸，并把 `kLast` 清掉。

手续费是在每笔 swap 里攒进 \(k\) 的，但协议的 LP 不是每笔 swap 都铸，而是拖到下一次添加或移除流动性时一次性结算。

## 白皮书的数字

最初存入 100 DAI 和 1 ETH，不考虑锁死的 1000 枚最小流动性，得到 \(S_1 = 10\)。过了一段时间，池子变成 96 DAI 和 1.5 ETH，中途没有别人存取。

$$
\sqrt{k_1} = \sqrt{100} = 10, \quad \sqrt{k_2} = \sqrt{96 \times 1.5} = \sqrt{144} = 12
$$

$$
S_m = \frac{12 - 10}{5 \times 12 + 10} \times 10 = \frac{20}{70} \approx 0.2857
$$

协议得到约 0.2857 个 LP。总供给变成约 10.2857，协议约占其中的 2.8%。