import { network } from "hardhat";

const { ethers } = await network.create({ network: "sepolia" });
const chain = await ethers.provider.getNetwork();
const [signer] = await ethers.getSigners();
const balance = await ethers.provider.getBalance(signer.address);

console.log("chainId", chain.chainId.toString());
console.log("address", signer.address);
console.log("balanceEth", ethers.formatEther(balance));
