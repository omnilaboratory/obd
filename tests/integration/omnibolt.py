import time

from typing import Optional, Final

import json
from websocket import create_connection

from omnicore_connection import OmnicoreConnection


class OmniBolt:
    def __init__(self, address_str: str):
        self._address_str = address_str
        self._websocket = None
        self._account = None
        self.user_id: Optional[str] = None
        self.node_id: Optional[str] = None
        self.node_address: Optional[str] = None
        self.public_key: Optional[str] = None
        self.wallet_address: Optional[str] = None
        self.wif: Optional[str] = None

    def _sent_sync_request(self, message, expected_type=None):

        if expected_type is None:
            expected_type = message["type"]

        str_message = json.dumps(message)
        print(str_message)
        self._websocket.send(str_message)

        raw_result = self._websocket.recv()

        result = json.loads(raw_result)

        assert (
            result["type"] == expected_type
        ), f"Expecting type={expected_type}, got type={result['type']}"

        print(result)
        assert result["status"], f"Invalid response=[{result}]"

        return result

    def connect(self):
        ws = create_connection(f"ws://{self._address_str}/wsregtest")
        open_message = '{"type": -102004}'
        ws.send(open_message)
        self._websocket = ws
        result = ws.recv()
        print("Received '%s'" % result)
        result = ws.recv()
        print("Received '%s'" % result)
        d = json.loads(result)
        return d["result"]

    def login(self, user) -> str:
        message = {"type": -102001, "data": {"mnemonic": user}}

        response = self._sent_sync_request(message)
        self.user_id = response["result"]["userPeerId"]
        self.node_id = response["result"]["nodePeerId"]
        self.node_address = response["result"]["nodeAddress"]

        return response

    def generate_wallet(self):
        message = {"type": -103000}
        response = self._sent_sync_request(message)
        self.public_key = response["result"]["pub_key"]
        self.wallet_address = response["result"]["address"]
        self.wif = response["result"]["wif"]
        return response

    def connect_peer(self, peer: "OmniBolt"):
        message = {"type": -102003, "data": {"remote_node_address": peer.node_address}}
        return self._sent_sync_request(message)

    def _open_channel(
        self,
        peer_node: "OmniBolt",
        funding_public_key,
    ):
        message = {
            "type": -100032,
            "data": {"funding_pubkey": funding_public_key},
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
        }
        print("Open channel request", message)
        response = self._sent_sync_request(message)
        response_peer = peer_node.read_message()

        return response, response_peer

    def accept_channel(
        self,
        peer_node: "OmniBolt",
        channel_id,
        funding_public_key,
        accept=True,
    ):
        message = {
            "type": -100033,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "funding_pubkey": funding_public_key,
                "approval": accept,
            },
        }

        reply = self._sent_sync_request(message)
        assert reply["status"], f"Invalid response=[{reply}]"

        reply_peer = peer_node.read_message()
        assert reply_peer["status"]

        return reply, reply_peer

    def funding_btc(
        self,
        from_address,
        from_address_private,
        to_address,
        amount=0.00012,
        miner_fee=0.00001,
    ):
        message = {
            "type": -102109,
            "data": {
                "from_address": from_address,
                "from_address_private_key": from_address_private,
                "to_address": to_address,
                "amount": amount,
                "miner_fee": miner_fee,
            },
        }

        return self._sent_sync_request(message)

    def sign_mining(self, peer_node: "OmniBolt", hex: str):
        message = {
            "type": -100341,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "hex": hex,
            },
        }
        reply = self._sent_sync_request(message)
        assert reply["status"], f"Invalid response=[{reply}]"

        reply_peer = peer_node.read_message()
        assert reply_peer["status"]

        return reply, reply_peer

    def sign_hex(self, hex: str, inputs):
        message = {
            "type": -102123,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {"hex": hex, "prvkey": self.wif, "inputs": inputs},
        }

        return self._sent_sync_request(message)

    def sign_hex_c1a(self, peer, hex: str):
        message = {
            "type": -101034,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {"signed_c1a_hex": hex},
        }

        response_alice = self._sent_sync_request(message)
        response_bob = peer.read_message()

        return response_alice, response_bob

    def btc_funding_created(
        self,
        peer_node: "OmniBolt",
        channel_id,
        channel_address_private_key,
        funding_tx_hex,
    ):
        message = {
            "type": -100340,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "channel_address_private_key": channel_address_private_key,
                "funding_tx_hex": funding_tx_hex,
            },
        }

        return self._sent_sync_request(message)

    def btc_funding_singed(
        self,
        peer_node: "OmniBolt",
        channel_id,
        channel_address_private_key,
        funding_tx_id: str,
        signed_miner_redeem_transaction_hex: str,
    ):
        message = {
            "type": -100350,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "channel_address_private_key": channel_address_private_key,
                "funding_txid": funding_tx_id,
                "signed_miner_redeem_transaction_hex": signed_miner_redeem_transaction_hex,
                "approval": True,
            },
        }

        reply = self._sent_sync_request(message)
        reply_peer = peer_node.read_message()

        return reply, reply_peer

    def funding_assert(
        self,
        from_address: str,
        to_address: str,
        amount: float,
        property_id: float,
    ):
        message = {
            "type": -102120,
            "data": {
                "from_address": from_address,
                "to_address": to_address,
                "amount": amount,
                "property_id": property_id,
            },
        }
        return self._sent_sync_request(message)

    def alice_signed_rd_of_asset_funding(self, channel_id: str, rd_signed_hex: str):
        message = {
            "type": -101134,
            "data": {
                "channel_id": channel_id,
                "rd_signed_hex": rd_signed_hex,
            },
        }
        return self._sent_sync_request(message)

    def signed_rd_and_br(
        self,
        peer_node: "OmniBolt",
        temporary_channel_id: str,
        br_id: int,
        rd_signed_hex: str,
        br_signed_hex: str,
    ):

        message = {
            "type": -101035,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": temporary_channel_id,
                "br_id": br_id,
                "rd_signed_hex": rd_signed_hex,
                "br_signed_hex": br_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer_node.read_message()

    def asset_funding_signed(
        self,
        peer_node: "OmniBolt",
        temporary_channel_id,
        signed_funding_assert_hex: str,
    ):
        message = {
            "type": -100035,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": temporary_channel_id,
                "signed_alice_rsmc_hex": signed_funding_assert_hex,
            },
        }

        return self._sent_sync_request(message)

    def asset_funding_created(
        self,
        peer_node: "OmniBolt",
        channel_id,
        funding_tx_hex,
        temp_address_public_key,
        temp_address_private_key,
        channel_address_priv,
    ):
        message = {
            "type": -100034,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "funding_tx_hex": funding_tx_hex,
                "temp_address_pub_key": temp_address_public_key,
                "temp_address_private_key": temp_address_private_key,
                "channel_address_private_key": channel_address_priv,
            },
        }
        reply = self._sent_sync_request(message)
        return reply

    def get_all_channels(self):
        message = {
            "type": -103150,
        }

        return self._sent_sync_request(message)["result"]["data"]

    def add_invoice(self, property: int, amount: int):

        lock = "abc"
        message = {
            "type": -100402,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {
                "h": str(hash(lock)),
                "expiry_time": "2020-12-15",
                "description": "description",
                "property_id": property,
                "amount": amount,
            },
        }

        return self._sent_sync_request(message)

    def _get_routing_information(self, invoice: str, peer: "OmniBolt"):
        message = {
            "type": -100401,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {"invoice": invoice},
        }

        return self._sent_sync_request(message)

    def _add_htlc(
        self,
        amount: float,
        h: str,
        memo: str,
        cltv_expiry: int,
        routing_packet: str,
        last_address_private_key: str,
        curr_rsmc_temp_address_index: int,
        curr_rsmc_temp_address_pub_key: str,
        curr_htlc_temp_address_for_ht1a_index: int,
        curr_htlc_temp_address_for_ht1a_pub_key: str,
        curr_htlc_temp_address_pub_key: str
    ):

        message = {
            "type": -100040,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {
                "amount": amount,
                "h": h,
                "cltv_expiry": cltv_expiry,
                "routing_packet": routing_packet,
                "last_temp_address_private_key": last_address_private_key,
                "curr_rsmc_temp_address_index": curr_rsmc_temp_address_index,
                "curr_rsmc_temp_address_pub_key": curr_rsmc_temp_address_pub_key,
                "curr_htlc_temp_address_pub_key": curr_htlc_temp_address_pub_key,
                "curr_htlc_temp_address_for_ht1a_index": curr_htlc_temp_address_for_ht1a_index,
                "curr_htlc_temp_address_for_ht1a_pub_key": curr_htlc_temp_address_for_ht1a_pub_key,
                "memo": memo,
            },
        }

        return self._sent_sync_request(message)


    def pay_invoice(self, invoice: str, peer: "OmniBolt"):
        routing_information = self._get_routing_information(invoice, peer)["result"]
        memo = routing_information["memo"]
        amount = routing_information["amount"]
        h = routing_information["h"]
        min_cltv_expiry = routing_information["min_cltv_expiry"]
        routing_packet = routing_information["routing_packet"]

        add_htlc_response = self._add_htlc(amount=amount,
                                           h=h,
                                           memo=memo,
                                           cltv_expiry=min_cltv_expiry,
                                           routing_packet=routing_packet,
                                           curr_rsmc_temp_address_index=1,
                                           curr_rsmc_temp_address_pub_key=self.public_key,
                                           curr_htlc_temp_address_for_ht1a_pub_key=self.public_key,
                                           curr_htlc_temp_address_for_ht1a_index=1,
                                           curr_htlc_temp_address_pub_key=self.public_key,
                                           last_address_private_key=self.wif)




    def read_message(self):
        raw_reply = self._websocket.recv()
        reply = json.loads(raw_reply)
        assert reply["status"], f"Invalid reply=[{reply}]"
        return reply

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
            }
        }

        return self._sent_sync_request(message)

    def open_channel(
        self,
        peer: "OmniBolt",
        omnicore_connection: OmnicoreConnection,
        property_id,
        funding_btc,
        channel_size: int,
    ):
        open_channel_resp, response_bob = self._open_channel(peer, self.public_key)

        temp_channel_id = open_channel_resp["result"]["temporary_channel_id"]

        accept_channel_response, funding_accept_reply = peer.accept_channel(
            self, temp_channel_id, peer.public_key
        )
        channel_address = funding_accept_reply["result"]["channel_address"]
        omnicore_connection.send_bitcoin(self.wallet_address, 100000)

        funding_count: Final = 3
        for _ in range(funding_count):
            funding_btc_response = self.funding_btc(
                self.wallet_address, self.wif, channel_address
            )

            funding_btc_txhex = funding_btc_response["result"]["hex"]

            signed_tx_hex = self.sign_hex(
                funding_btc_txhex, funding_btc_response["result"]["inputs"]
            )["result"]

            omnicore_connection.mine_bitcoin(20, funding_btc)

            funding_created_response = self.btc_funding_created(
                peer,
                temp_channel_id,
                self.wif,
                signed_tx_hex,
            )

            funding_created_response_hex = self.sign_hex(
                funding_created_response["result"]["hex"],
                funding_created_response["result"]["inputs"],
            )["result"]
            response_sign_mine, response_sign_mine_bob = self.sign_mining(
                peer, funding_created_response_hex
            )

            response_sign_mine_signed = peer.sign_hex(
                response_sign_mine["result"]["sign_data"]["hex"],
                response_sign_mine["result"]["sign_data"]["inputs"],
            )["result"]

            btc_funding_created_response = self.btc_funding_created(
                peer,
                temp_channel_id,
                self.wif,
                signed_tx_hex,
            )

            peer.read_message()

            peer.btc_funding_singed(
                self,
                temp_channel_id,
                peer.wif,
                btc_funding_created_response["result"]["funding_txid"],
                response_sign_mine_signed,
            )
        funding_asset_response = self.funding_assert(
            self.wallet_address,
            channel_address,
            channel_size,
            property_id,
        )
        funding_assert_hex = funding_asset_response["result"]["hex"]
        signed_funding_assert_hex = self.sign_hex(
            funding_assert_hex, funding_asset_response["result"]["inputs"]
        )["result"]

        omnicore_connection.mine_bitcoin(20, funding_btc)
        funding_created_response = self.asset_funding_created(
            peer,
            temp_channel_id,
            signed_funding_assert_hex,
            self.public_key,
            self.wif,
            self.wif,
        )

        funding_created_response_hex_signed_alice = self.sign_hex(
            funding_created_response["result"]["hex"],
            funding_created_response["result"]["inputs"],
        )["result"]

        response_sign_c1a, response_bob = self.sign_hex_c1a(
            peer, funding_created_response_hex_signed_alice
        )

        response_signed_again = peer.sign_hex(
            response_sign_c1a["result"]["sign_data"]["hex"],
            response_sign_c1a["result"]["sign_data"]["inputs"],
        )["result"]

        response_bob = peer.asset_funding_signed(
            self, temp_channel_id, response_signed_again
        )["result"]
        alice_sign_data_br = response_bob["alice_br_sign_data"]
        alice_sign_data_rd = response_bob["alice_rd_sign_data"]

        rd_bob_signed = peer.sign_hex(
            alice_sign_data_rd["hex"], alice_sign_data_rd["inputs"]
        )["result"]

        bob_sign_response, bob_sign_response_alice = peer.signed_rd_and_br(
            peer_node=self,
            temporary_channel_id=temp_channel_id,
            br_id=alice_sign_data_br["br_id"],
            br_signed_hex=peer.sign_hex(
                alice_sign_data_br["hex"], alice_sign_data_br["inputs"]
            )["result"],
            rd_signed_hex=rd_bob_signed,
        )
        channel_id = bob_sign_response["result"]["channel_id"]
        rd_alice_signed = self.sign_hex(rd_bob_signed, alice_sign_data_rd["inputs"])[
            "result"
        ]

        self.alice_signed_rd_of_asset_funding(
            channel_id, rd_alice_signed
        )

        omnicore_connection.mine_bitcoin(20, funding_btc)
        return channel_id
