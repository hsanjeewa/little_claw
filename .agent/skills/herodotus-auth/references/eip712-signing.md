# EIP-712 Signing

The challenge endpoint returns the full typed-data payload. Sign those fields exactly; do not rebuild the domain, message, primary type, or statement locally.

## viem

```ts
const signature = await account.signTypedData({
  domain: challenge.eip712.domain,
  types: challenge.eip712.types,
  primaryType: challenge.eip712.primaryType,
  message: challenge.eip712.message,
});
```

## ethers v6

`ethers` expects the domain type to be omitted from the `types` object.

```ts
const { EIP712Domain: _drop, ...types } = challenge.eip712.types;
const signature = await wallet.signTypedData(
  challenge.eip712.domain,
  types,
  challenge.eip712.message,
);
```

## Browser Wallet

```ts
const signature = await window.ethereum.request({
  method: 'eth_signTypedData_v4',
  params: [address, JSON.stringify(challenge.eip712)],
});
```

## KMS or Hardware

Use the returned typed-data payload as the canonical message, convert it with your signer library's EIP-712 encoder, and return a normal 65-byte EVM signature. Keep the Herodotus auth exchange outside the signer so the signer only sees typed data and never sees Herodotus bearer tokens.
