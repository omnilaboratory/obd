from typing import Optional

from omnibolt import OmniBolt
from omnicore_connection import OmnicoreConnection
import os


class TestRunner:
    @staticmethod
    def generate_omni_bolt(address: str) -> OmniBolt:
        omni_bolt = OmniBolt(address)
        user = omni_bolt.connect()
        omni_bolt.login(user)
        omni_bolt.wallet_address = omni_bolt.generate_address()

        return omni_bolt

    def __init__(self):
        self.omni_bolt_alice = self.generate_omni_bolt(
            os.environ.get("OMNI_BOLT_ALICE")
        )
        self.omni_bolt_bob = self.generate_omni_bolt(os.environ.get("OMNI_BOLT_BOB"))
        self.property_id: Optional[int] = None
        self.channel_id: Optional[int] = None

    def setup_basic_workflow(self, channel_size: int) -> int:
        omnicore_connection = OmnicoreConnection()

        address_miner = omnicore_connection.generate_bitcoin_address("miner")
        address_master_funder = omnicore_connection.generate_bitcoin_address(
            "address_master_funder"
        )

        omnicore_connection.mine_bitcoin(200, address_miner)
        omnicore_connection.send_bitcoin(address_master_funder, 100000)
        created_omnicore_item = omnicore_connection.generate_omni_token(
            address_master_funder, address_miner
        )

        self.omni_bolt_alice.connect_peer(
            self.omni_bolt_bob,
        ),

        grant_amount = "9999999999.00000000"

        omnicore_connection.omni_sendgrant(
            address_master_funder,
            created_omnicore_item,
            grant_amount,
        )

        # Send omnicore currency to Alice

        omnicore_connection.mine_bitcoin(20, address_miner)
        self.omni_bolt_alice.property_id = int(created_omnicore_item["propertyid"])
        self.omni_bolt_bob.property_id = int(created_omnicore_item["propertyid"])
        self.omni_bolt_alice.open_channel(
            self.omni_bolt_bob,
            omnicore_connection,
            created_omnicore_item,
            address_master_funder,
            channel_size,
        )
