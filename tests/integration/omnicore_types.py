from dataclasses import dataclass


@dataclass(frozen=False)
class Address:
    public_key: str
    private_key: str
    address: str
    index: int
