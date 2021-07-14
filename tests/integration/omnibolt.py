from typing import Optional

import json
from websocket import create_connection

from omnicore_connection import OmnicoreConnection


from omnicore_types import Address


class OmniBolt:
    def __init__(self, address_str: str):
        self._address_str = address_str
        self._websocket = None
        self._account = None
        self.user_id: Optional[str] = None
        self.node_id: Optional[str] = None
        self.wallet_address: Optional[Address] = None

    def _sent_sync_request(self, message, expected_type=None):

        attempts = 0
        max_attempts = 10
        while attempts < max_attempts:

            if expected_type is None:
                expected_type = message["type"]

            str_message = json.dumps(message)
            print(str_message)
            self._websocket.send(str_message)

            raw_result = self._websocket.recv()

            result = json.loads(raw_result)

            try:
                response = result['result']

                assert (
                    result["type"] == expected_type
                ), f"Expecting type={expected_type}, got type={result['type'], }"

                assert result["status"], f"Invalid response=[{result}]"
                return result["result"]
            except KeyError:
                attempts = attempts + 1

        raise RuntimeError("Can't get a valid response")

    def connect(self):
        attempt = 0
        max_attempts = 10
        while attempt < max_attempts:
            ws = create_connection(f"ws://{self._address_str}/wsregtest")
            open_message = '{"type": -102004}'
            ws.send(open_message)
            self._websocket = ws
            result = ws.recv()
            print("Received '%s'" % result)
            result = ws.recv()
            print("Received '%s'" % result)
            d = json.loads(result)
            try:
                return d["result"]
            except KeyError:
                print(f"Invalid response {d}")
                attempt = attempt + 1

    def login(self, user) -> str:
        message = {"type": -102001, "data": {"mnemonic": user, "is_admin": True}}

        response = self._sent_sync_request(message)
        self.user_id = response["userPeerId"]
        self.node_id = response["nodePeerId"]
        self.node_address = response["nodeAddress"]

        return response

    def generate_address(
        self,
    ) -> Address:
        message = {"type": -103000}
        response = self._sent_sync_request(message)
        return Address(
            public_key=response["pub_key"],
            private_key=response["wif"],
            index=response["index"],
            address=response["address"],
        )

    def connect_peer(self, peer: "OmniBolt"):
        message = {"type": -102003, "data": {"remote_node_address": peer.node_address}}
        return self._sent_sync_request(message)

    def _open_channel(
        self,
        *,
        peer_node: "OmniBolt",
        funding_public_key: str,
        funder_address_index: int,
    ):
        message = {
            "type": -100032,
            "data": {
                "funding_pubkey": funding_public_key,
                "funder_address_index": funder_address_index,
            },
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
        }
        response = self._sent_sync_request(message)
        return response, self.read_message()

    def auto_funding(
        self,
        *,
        peer: "OmniBolt",
        temporary_channel_id: str,
        btc_amount: float,
        property_id: int,
        asset_amount: float,
    ):
        message = {
            "type": -100134,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "temporary_channel_id": temporary_channel_id,
                "btc_amount": btc_amount,
                "property_id": property_id,
                "asset_amount": asset_amount,
            },
        }
        return (
            self._sent_sync_request(message, expected_type=-100340),
            self.read_message(),
            self.read_message(),
            self.read_message(),
            self.read_message(),
            self.read_message(),
            self.read_message(),
        )

    def get_all_channels(self):
        message = {
            "type": -103150,
        }

        return self._sent_sync_request(message)["data"]

    def add_invoice(self, property_id: int, amount: int):

        response = self.generate_address()

        message = {
            "type": -100402,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {
                "h": response.public_key,
                "expiry_time": "2120-12-15",
                "description": "description",
                "property_id": property_id,
                "amount": amount,
            },
        }

        self.h_address = response

        return self._sent_sync_request(message)

    def _get_routing_information(self, invoice: str, peer: "OmniBolt"):
        message = {
            "type": -100401,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {"invoice": invoice},
        }

        return (
            self._sent_sync_request(message, expected_type=-100040),
            self.read_message(),
            self.read_message(),
            self.read_message(),
            self.read_message(),
        )

    def pay_invoice(self, invoice: str, peer: "OmniBolt"):
        self._get_routing_information(invoice, peer), peer.read_message()

    def read_message(self):
        raw_reply = self._websocket.recv()
        reply = json.loads(raw_reply)
        assert reply["status"], f"Invalid reply=[{reply}]"
        return reply["result"]

    def create_commitment_transaction(
        self,
        peer: "OmniBolt",
        channel_id: str,
        amount: float,
        last_temp_address_private_key: str,
        curr_temp_address_pub_key: str,
        curr_temp_address_index: int,
    ):

        message = {
            "type": -100351,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "amount": amount,
                "last_temp_address_private_key": last_temp_address_private_key,
                "curr_temp_address_index": curr_temp_address_index,
                "curr_temp_address_pub_key": curr_temp_address_pub_key,
            },
        }

        return self._sent_sync_request(message)

    def open_channel(
        self,
        peer: "OmniBolt",
        omnicore_connection: OmnicoreConnection,
        omnicore_item,
        funding_btc,
        channel_size: int,
    ):
        property_id = omnicore_item["propertyid"]

        funding_address = self.generate_address()
        open_channel_resp, accept_message = self._open_channel(
            peer_node=peer,
            funding_public_key=funding_address.public_key,
            funder_address_index=funding_address.index,
        )

        temp_channel_id = open_channel_resp["temporary_channel_id"]

        omnicore_connection.send_bitcoin(funding_address.address, 1000000000)
        omnicore_connection.mine_bitcoin(50, self.wallet_address.address)

        grant_amount = "9999999999.00000000"
        omnicore_connection.send_omnicore_token(
            funding_btc,
            funding_address.address,
            omnicore_item,
            grant_amount,
        )

        omnicore_connection.mine_bitcoin(200, self.wallet_address.address)

        return self.auto_funding(
            temporary_channel_id=temp_channel_id,
            peer=peer,
            btc_amount=0.0004,
            property_id=property_id,
            asset_amount=channel_size,
        )
