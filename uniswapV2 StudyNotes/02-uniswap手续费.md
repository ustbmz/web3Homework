Uniswap V2 的手续费分两层。每笔交易先收 0.30%，这笔钱留在池子里，所以 \(k = x \cdot y\) 变大，流动性提供者一起受益。工厂如果设置了 `feeTo`，协议再从这笔增长里拿走 \(\dfrac{1}{6}\)，相当于交易额的 0.05%，剩下 \(\dfrac{5}{6}\) 仍归流动性提供者。

## 交易费怎么让 \(k\) 变大

用一种币换另一种币时，先从打进池子的数量里扣 0.30%，只用剩下的 99.7% 去推恒定乘积。代码里就是输入乘 997，再除以 1000：

$$
\text{amountOut} = \frac{997 \cdot \text{amountIn} \cdot \text{reserveOut}}{1000 \cdot \text{reserveIn} + 997 \cdot \text{amountIn}}
$$

扣下的 0.30% 没有转走，还在池子里。换完之后，两种币真实余额的乘积大于换之前的 \(k\)。

## 流动性提供者怎么受益

记换之前是 \(k_1\)，储备更新之后是 \(k_2\)。池子深度 \(L = \sqrt{k}\) 的相对增长是：

$$
f_{1,2} = 1 - \sqrt{\frac{k_1}{k_2}} = \frac{\sqrt{k_2} - \sqrt{k_1}}{\sqrt{k_2}}
$$

LP 的数量没变，每个 LP 对应的池子变厚了。撤出时按份额取回，两种币都会比没有手续费时多一点。

## 协议拿走六分之一

\(\phi\) 是协议要从这笔增长里分走的比例，V2 固定为 \(\dfrac{1}{6}\)。

从 \(t_1\) 到 \(t_2\)，\(\sqrt{k}\) 从 \(\sqrt{k_1}\) 涨到 \(\sqrt{k_2}\)，当时已有的 LP 总量是 \(S_1\)。铸给 `feeTo` 的数量是：

$$
S_m = \frac{\sqrt{k_2} - \sqrt{k_1}}{\left(\dfrac{1}{\phi} - 1\right)\sqrt{k_2} + \sqrt{k_1}} \cdot S_1
$$

代入 \(\phi = \dfrac{1}{6}\)，得到 \(\dfrac{1}{\phi} - 1 = 5\)：

$$
S_m = \frac{\sqrt{k_2} - \sqrt{k_1}}{5\sqrt{k_2} + \sqrt{k_1}} \cdot S_1
$$

分母里的 5 就是流动性提供者留 5 份，协议拿 1 份。

## 对应到合约

`mint` 和 `burn` 一开始都先调用 `_mintFee`。它比较的是现在的 \(\sqrt{\text{reserve}_0 \cdot \text{reserve}_1}\)，和上次流动性变动时记下的 \(\sqrt{k_{\text{Last}}}\)。

$$
\text{liquidity} = \frac{\text{totalSupply} \cdot (\sqrt{k} - \sqrt{k_{\text{Last}}})}{5\sqrt{k} + \sqrt{k_{\text{Last}}}}
$$

这就是上面的 \(S_m\)。铸完后，这次添加或移除结束时，把新的 \(k\) 写入 `kLast`。`feeTo` 是零地址时不铸，并把 `kLast` 清掉。

手续费在每笔 swap 里攒进 \(k\)，协议的 LP 不是每笔 swap 都铸，而是拖到下一次添加或移除流动性时一次性结算。

## 白皮书的数字

最初存入 100 DAI 和 1 ETH。不考虑锁死的 1000 枚最小流动性，得到 \(S_1 = 10\)。过了一段时间，池子变成 96 DAI 和 1.5 ETH，中途没有别人存取。

$$
\sqrt{k_1} = \sqrt{100} = 10
$$

$$
\sqrt{k_2} = \sqrt{96 \times 1.5} = \sqrt{144} = 12
$$

$$
S_m = \frac{12 - 10}{5 \times 12 + 10} \times 10 = \frac{20}{70} \approx 0.2857
$$

协议得到约 0.2857 个 LP。总供给变成约 10.2857，协议约占其中的 2.8%。