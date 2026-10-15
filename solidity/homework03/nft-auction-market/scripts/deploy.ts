import { mkdir, writeFile } from "node:fs/promises";
import { network } from "hardhat";

// Sepolia 上的 Chainlink 代理地址。部署前再对一下官方文档：
// https://docs.chain.link/data-feeds/price-feeds/addresses
const SEPOLIA_ETH_USD = "0x694AA1769357215DE4FAC081bf1f309aDC325306";
const SEPOLIA_LINK_USD = "0xc59E3633BAAC79493d908e63626716e204A45EdF";
const SEPOLIA_LINK = "0x779877A7B0D9E8603169DdbD7836e478b4624789";

const useSepolia = process.argv.includes("sepolia");

const { ethers } = await network.create(
  useSepolia ? { network: "sepolia" } : { network: "hardhatMainnet", chainType: "l1" },
);

const [deployer] = await ethers.getSigners();

let ethFeed: string;
let erc20: string;
let erc20Feed: string;

if (useSepolia) {
  ethFeed = SEPOLIA_ETH_USD;
  erc20 = SEPOLIA_LINK;
  erc20Feed = SEPOLIA_LINK_USD;
} else {
  // 本地没有 Chainlink，部署同样接口的模拟价格源。1 ETH = 2000 美元，1 USDC = 1 美元。
  const ethAggregator = await ethers.deployContract("MockV3Aggregator", [8, 2000n * 10n ** 8n]);
  const erc20Aggregator = await ethers.deployContract("MockV3Aggregator", [8, 1n * 10n ** 8n]);
  const token = await ethers.deployContract("MockERC20", ["Mock USDC", "USDC", 6]);
  await Promise.all([ethAggregator.waitForDeployment(), erc20Aggregator.waitForDeployment(), token.waitForDeployment()]);
  ethFeed = await ethAggregator.getAddress();
  erc20Feed = await erc20Aggregator.getAddress();
  erc20 = await token.getAddress();
}

// Sepolia 不会立刻出块。代理构造时会检查实现合约地址上是否已有代码，
// 所以每一笔部署都要等交易进块，再发下一笔。
const nft = await ethers.deployContract("AuctionNFT", [deployer.address]);
await nft.waitForDeployment();
console.log("AuctionNFT", await nft.getAddress());

const implementation = await ethers.deployContract("NftAuction");
await implementation.waitForDeployment();
console.log("NftAuction 实现合约", await implementation.getAddress());

const initData = implementation.interface.encodeFunctionData("initialize", [
  deployer.address,
  ethFeed,
  deployer.address,
  24n * 60n * 60n,
]);
const proxy = await ethers.deployContract("AuctionProxy", [await implementation.getAddress(), initData]);
await proxy.waitForDeployment();
console.log("拍卖代理", await proxy.getAddress());

const auction = await ethers.getContractAt("NftAuction", await proxy.getAddress());
const feedTx = await auction.setPriceFeed(erc20, erc20Feed);
await feedTx.wait();

const addresses = {
  network: useSepolia ? "sepolia" : "hardhat",
  deployer: deployer.address,
  nft: await nft.getAddress(),
  auctionProxy: await proxy.getAddress(),
  auctionImplementation: await implementation.getAddress(),
  ethUsdFeed: ethFeed,
  erc20,
  erc20UsdFeed: erc20Feed,
};

await mkdir("deployments", { recursive: true });
const file = useSepolia ? "deployments/sepolia.json" : "deployments/local.json";
await writeFile(file, `${JSON.stringify(addresses, null, 2)}\n`);

console.log(`已写入 ${file}`);
console.log(addresses);
console.log("用户要调用的拍卖地址是 auctionProxy。授权 NFT 时也授权这个地址。");
