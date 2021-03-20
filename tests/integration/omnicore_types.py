from dataclasses import dataclass
from typing import Optional


@dataclass(frozen=False)
class Address:
    public_key: str
    private_key: str
    address: str
    index: int


@dataclass
class ChannelAddresses:
    htlc_address: Address
    htlc_ht1a_address: Address
    rmsc_address: Address
    he_address: Address
    temp_address: Address
