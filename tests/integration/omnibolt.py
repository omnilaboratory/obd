from bit.network.meta import Unspent

from typing import Optional, Final, Any

import json
from websocket import create_connection

from omnicore_connection import OmnicoreConnection

from bit import wif_to_key
from bit import MultiSigTestnet

from omnicore_types import Address
from omnicore_types import ChannelAddresses


class OmniBolt:
    def __init__(self, address_str: str):
        self._address_str = address_str
        self._websocket = None
        self._account = None
        self.user_id: Optional[str] = None
        self.node_id: Optional[str] = None
        self.wallet_address: Optional[Address] = None
        self.channel_addresses: dict[str, ChannelAddresses] = {}

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
        return result["result"]

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
        message = {"type": -102001, "data": {"mnemonic": user, "is_admin": False}}

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

    def generate_address_multisig(self, pub_keys: list[str], sign_count=2):
        message = {
            "type": -102110,
            "data": {"sign_count": sign_count, "pub_keys": pub_keys},
        }
        response = self._sent_sync_request(message)
        return response

    def connect_peer(self, peer: "OmniBolt"):
        message = {"type": -102003, "data": {"remote_node_address": peer.node_address}}
        return self._sent_sync_request(message)

    def _open_channel(
        self,
        *,
        peer_node: "OmniBolt",
        funding_public_key,
    ):
        message = {
            "type": -100032,
            "data": {"funding_pubkey": funding_public_key},
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
        }
        response = self._sent_sync_request(message)
        return response, peer_node.read_message()

    def accept_channel(
        self,
        *,
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
        reply_peer = peer_node.read_message()
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

    def sign_mining(self, peer_node: "OmniBolt", hex_val: str):
        message = {
            "type": -100341,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "hex": hex_val,
            },
        }
        reply = self._sent_sync_request(message)
        reply_peer = peer_node.read_message()

        return reply, reply_peer

    @staticmethod
    def sign_hex(hex_val: str, inputs, private_key):

        key = wif_to_key(private_key)
        return key.sign_transaction(
            tx_data=hex_val,
            unspents=[
                Unspent(
                    amount=input_["amount"],
                    confirmations=0,
                    script=input_["scriptPubKey"],
                    txid=input_["txid"],
                    txindex=input_["vout"],
                )
                for input_ in inputs
            ],
        )

    @staticmethod
    def sign_hex_p2skh(hex_val: str, inputs, pub_keys, private_key):

        key = wif_to_key(private_key)

        try:
            multi_sig = MultiSigTestnet(key, pub_keys, m=2)
        except Exception as e:
            print("Error {}".format(e))
            raise e

        unspents = [
            Unspent(
                amount=input_["amount"],
                confirmations=0,
                script=input_["scriptPubKey"],
                txid=input_["txid"],
                txindex=input_["vout"],
            )
            for input_ in inputs
        ]

        try:
            return multi_sig.sign_transaction(tx_data=hex_val, unspents=unspents)
        except Exception as e:
            print("Error {}".format(e))
            raise e

    def sign_hex_c1a(self, peer, hex_val: str):
        message = {
            "type": -101034,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {"signed_c1a_hex": hex_val},
        }

        response_alice = self._sent_sync_request(message)
        response_bob = peer.read_message()

        return response_alice, response_bob

    def btc_funding_created(
        self,
        *,
        peer_node: "OmniBolt",
        channel_id,
        funding_tx_hex,
    ):
        message = {
            "type": -100340,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "funding_tx_hex": funding_tx_hex,
            },
        }

        return self._sent_sync_request(message)

    def btc_funding_singed(
        self,
        *,
        peer_node: "OmniBolt",
        channel_id,
        funding_tx_id: str,
        signed_miner_redeem_transaction_hex: str,
    ):
        message = {
            "type": -100350,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
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

    def alice_signed_rd_of_asset_funding(self, *, channel_id: str, rd_signed_hex: str):
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
        *,
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
        *,
        peer_node: "OmniBolt",
        channel_id: str,
        funding_tx_hex: str,
        temp_address_public_key: str,
        temp_address_index: int,
    ):
        message = {
            "type": -100034,
            "recipient_node_peer_id": peer_node.node_id,
            "recipient_user_peer_id": peer_node.user_id,
            "data": {
                "temporary_channel_id": channel_id,
                "funding_tx_hex": funding_tx_hex,
                "temp_address_pub_key": temp_address_public_key,
                "temp_address_index": temp_address_index,
            },
        }
        reply = self._sent_sync_request(message)
        return reply

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

        return self._sent_sync_request(message)

    def _add_htlc(
        self,
        peer: "OmniBolt",
        amount: float,
        h: str,
        memo: str,
        cltv_expiry: int,
        routing_packet: str,
        last_temp_address_private_key: str,
        curr_rsmc_temp_address_index: int,
        curr_rsmc_temp_address_pub_key: str,
        curr_htlc_temp_address_for_ht1a_index: int,
        curr_htlc_temp_address_for_ht1a_pub_key: str,
        curr_htlc_temp_address_pub_key: str,
    ):

        message = {
            "type": -100040,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "amount": amount,
                "amount_to_payee": amount,
                "h": h,
                "cltv_expiry": cltv_expiry,
                "routing_packet": routing_packet,
                "last_temp_address_private_key": last_temp_address_private_key,
                "curr_rsmc_temp_address_index": curr_rsmc_temp_address_index,
                "curr_rsmc_temp_address_pub_key": curr_rsmc_temp_address_pub_key,
                "curr_htlc_temp_address_pub_key": curr_htlc_temp_address_pub_key,
                "curr_htlc_temp_address_for_ht1a_index": curr_htlc_temp_address_for_ht1a_index,
                "curr_htlc_temp_address_for_ht1a_pub_key": curr_htlc_temp_address_for_ht1a_pub_key,
                "memo": memo,
                "is_pay_invoice": True,
            },
        }

        return self._sent_sync_request(message)

    def _sign_inputs_p2skh(
        self: "OmniBolt", input_val, private_key, pub_key_override=None
    ):

        pub_keys = pub_key_override or [
            input_val["pub_key_a"],
            input_val["pub_key_b"],
        ]

        return self.sign_hex_p2skh(
            input_val["hex"], input_val["inputs"], pub_keys, private_key=private_key
        )

    def alice_sign_c3a(
        self,
        peer: "OmniBolt",
        channel_id: str,
        rsmc_partial_signed_hex: str,
        counterparty_partial_signed_hex: str,
        htlc_partial_signed_hex: str,
    ):

        message = {
            "type": -100100,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3a_rsmc_partial_signed_hex": rsmc_partial_signed_hex,
                "c3a_counterparty_partial_signed_hex": counterparty_partial_signed_hex,
                "c3a_htlc_partial_signed_hex": htlc_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def sign_c3b_bob_side(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c3a_rsmc_rd_partial_signed_hex: str,
        c3a_rsmc_br_partial_signed_hex: str,
        c3a_htlc_ht_partial_signed_hex: str,
        c3a_htlc_hlock_partial_signed_hex: str,
        c3a_htlc_br_partial_signed_hex: str,
        c3b_rsmc_partial_signed_hex: str,
        c3b_counterparty_partial_signed_hex: str,
        c3b_htlc_partial_signed_hex: str,
    ):

        message = {
            "type": -100101,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3a_rsmc_rd_partial_signed_hex": c3a_rsmc_rd_partial_signed_hex,
                "c3a_rsmc_br_partial_signed_hex": c3a_rsmc_br_partial_signed_hex,
                "c3a_htlc_ht_partial_signed_hex": c3a_htlc_ht_partial_signed_hex,
                "c3a_htlc_hlock_partial_signed_hex": c3a_htlc_hlock_partial_signed_hex,
                "c3a_htlc_br_partial_signed_hex": c3a_htlc_br_partial_signed_hex,
                "c3b_rsmc_partial_signed_hex": c3b_rsmc_partial_signed_hex,
                "c3b_counterparty_partial_signed_hex": c3b_counterparty_partial_signed_hex,
                "c3b_htlc_partial_signed_hex": c3b_htlc_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def sign_bob_side(
        self,
        payer_commitment_tx_hash: str,
        c3a_complete_signed_rsmc_hex: str,
        c3a_complete_signed_counterparty_hex: str,
        c3a_complete_signed_htlc_hex: str,
        last_temp_address_private_key: str,
        curr_rsmc_temp_address_index: int,
        curr_rsmc_temp_address_pub_key: str,
        curr_htlc_temp_address_index: int,
        curr_htlc_temp_address_pub_key: str,
    ):

        message = {
            "type": -100041,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {
                "payer_commitment_tx_hash": payer_commitment_tx_hash,
                "c3a_complete_signed_rsmc_hex": c3a_complete_signed_rsmc_hex,
                "c3a_complete_signed_counterparty_hex": c3a_complete_signed_counterparty_hex,
                "c3a_complete_signed_htlc_hex": c3a_complete_signed_htlc_hex,
                "last_temp_address_private_key": last_temp_address_private_key,
                "curr_rsmc_temp_address_index": curr_rsmc_temp_address_index,
                "curr_rsmc_temp_address_pub_key": curr_rsmc_temp_address_pub_key,
                "curr_htlc_temp_address_index": curr_htlc_temp_address_index,
                "curr_htlc_temp_address_pub_key": curr_htlc_temp_address_pub_key,
            },
        }

        return self._sent_sync_request(message)

    def _htlc_close_curr_tx(
        self,
        peer: "OmniBolt",
        channel_id: str,
        last_rsmc_temp_address_private_key: str,
        last_htlc_temp_address_private_key: str,
        last_htlc_temp_address_for_htnx_private_key: str,
        curr_temp_address_index: int,
        curr_temp_address_pub_key: str,
    ):

        message = {
            "type": -100049,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "last_rsmc_temp_address_private_key": last_rsmc_temp_address_private_key,
                "last_htlc_temp_address_private_key": last_htlc_temp_address_private_key,
                "last_htlc_temp_address_for_htnx_private_key": last_htlc_temp_address_for_htnx_private_key,
                "curr_temp_address_index": curr_temp_address_index,
                "curr_temp_address_pub_key": curr_temp_address_pub_key,
            },
        }

        return self._sent_sync_request(message)

    def _alice_sign_rsmc_for_c4a(
        self,
        peer: "OmniBolt",
        channel_id: str,
        rsmc_partial_signed_hex: str,
        counterparty_partial_signed_hex: str,
    ):

        message = {
            "type": -100110,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "rsmc_partial_signed_hex": rsmc_partial_signed_hex,
                "counterparty_partial_signed_hex": counterparty_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _bob_sign_close_htlc(
        self,
        peer: "OmniBolt",
        msg_hash: str,
        c4a_rsmc_complete_signed_hex: str,
        last_rsmc_temp_address_private_key: str,
        last_htlc_temp_address_private_key: str,
        last_htlc_temp_address_for_htnx_private_key: str,
        curr_temp_address_index: int,
        curr_temp_address_pub_key: str,
        c4a_counterparty_complete_signed_hex: str,
    ):

        message = {
            "type": -100050,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "msg_hash": msg_hash,
                "c4a_rsmc_complete_signed_hex": c4a_rsmc_complete_signed_hex,
                "c4a_counterparty_complete_signed_hex": c4a_counterparty_complete_signed_hex,
                "last_rsmc_temp_address_private_key": last_rsmc_temp_address_private_key,
                "last_htlc_temp_address_private_key": last_htlc_temp_address_private_key,
                "last_htlc_temp_address_for_htnx_private_key": last_htlc_temp_address_for_htnx_private_key,
                "curr_temp_address_index": curr_temp_address_index,
                "curr_temp_address_pub_key": curr_temp_address_pub_key,
            },
        }

        return self._sent_sync_request(message)

    def _bob_signed_rsmc_for_c4b(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c4a_rd_signed_hex: str,
        c4a_br_signed_hex: str,
        c4a_br_id: int,
        c4b_rsmc_signed_hex: str,
        c4b_counterparty_signed_hex: str,
    ):

        message = {
            "type": -100111,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c4a_rd_signed_hex": c4a_rd_signed_hex,
                "c4a_br_signed_hex": c4a_br_signed_hex,
                "c4a_br_id": c4a_br_id,
                "c4b_rsmc_signed_hex": c4b_rsmc_signed_hex,
                "c4b_counterparty_signed_hex": c4b_counterparty_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _send_back_r(self, peer: "OmniBolt", channel_id: str, r: str):

        message = {
            "type": -100045,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "r": r,
                "channel_id": channel_id,
            },
        }

        return self._sent_sync_request(message)

    def _bob_sign_herd_for_csb(
        self, peer: "OmniBolt", channel_id: str, c3b_htlc_herd_partial_signed_hex: str
    ):

        message = {
            "type": -100106,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3b_htlc_herd_partial_signed_hex": c3b_htlc_herd_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _alice_sign_herd(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c3b_htlc_herd_complete_signed_hex: str,
        c3b_htlc_hebr_partial_signed_hex: str,
    ):
        message = {
            "type": -100046,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3b_htlc_herd_complete_signed_hex": c3b_htlc_herd_complete_signed_hex,
                "c3b_htlc_hebr_partial_signed_hex": c3b_htlc_hebr_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _on_get_herd_at_bob(
        self, peer: "OmniBolt", channel_id: str, c3b_htlc_herd_complete_signed_hex: str
    ):

        message = {
            "type": -110046,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3b_htlc_herd_complete_signed_hex": c3b_htlc_herd_complete_signed_hex,
            },
        }

        return self._sent_sync_request(message)

    def _on_alice_signed_cxb(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c4a_rd_complete_signed_hex: str,
        c4b_rsmc_complete_signed_hex: str,
        c4b_counterparty_complete_signed_hex: str,
    ):

        message = {
            "type": -100112,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c4a_rd_complete_signed_hex": c4a_rd_complete_signed_hex,
                "c4b_rsmc_complete_signed_hex": c4b_rsmc_complete_signed_hex,
                "c4b_counterparty_complete_signed_hex": c4b_counterparty_complete_signed_hex,
            },
        }

        return self._sent_sync_request(message)

    def _on_alice_signed_c4b(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c4b_rd_partial_signed_hex: str,
        c4b_br_partial_signed_hex: str,
        c4b_br_id: int,
    ):

        message = {
            "type": -100113,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c4b_rd_partial_signed_hex": c4b_rd_partial_signed_hex,
                "c4b_br_partial_signed_hex": c4b_br_partial_signed_hex,
                "c4b_br_id": c4b_br_id,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _bob_signed_cxb_sub_tx(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c4b_rd_complete_signed_hex: str,
    ):
        message = {
            "type": -100114,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c4b_rd_complete_signed_hex": c4b_rd_complete_signed_hex,
            },
        }

        return self._sent_sync_request(message)

    def backwards_workflow(self, peer: "OmniBolt", channel_id: str, r: str):
        address_peer = peer.channel_addresses[channel_id]

        resp_send_back_r = self._send_back_r(peer, channel_id, r)
        resp_bob_sign_herd, resp_alice_sign_herd = self._bob_sign_herd_for_csb(
            peer,
            channel_id=channel_id,
            c3b_htlc_herd_partial_signed_hex=self._sign_inputs_p2skh(
                resp_send_back_r["c3b_htlc_herd_raw_data"],
                address_peer.he_address.private_key,
            ),
        )

        resp_alice_sign_herd, resp_alice_sign_herd_bob = peer._alice_sign_herd(
            peer=self,
            channel_id=channel_id,
            c3b_htlc_hebr_partial_signed_hex=peer._sign_inputs_p2skh(
                resp_alice_sign_herd["c3b_htlc_hebr_raw_data"],
                address_peer.he_address.private_key,
            ),
            c3b_htlc_herd_complete_signed_hex=peer._sign_inputs_p2skh(
                resp_alice_sign_herd["c3b_htlc_herd_partial_signed_data"],
                address_peer.he_address.private_key,
            ),
        )
        return resp_alice_sign_herd, resp_alice_sign_herd_bob

    def close_workflow(
        self,
        peer: "OmniBolt",
        channel_id: str,
    ):

        self.channel_addresses[channel_id].temp_address = self.generate_address()
        peer.channel_addresses[channel_id].temp_address = self.generate_address()

        address_self = self.channel_addresses[channel_id]
        address_peer = peer.channel_addresses[channel_id]

        funding_address_self = self.funding_address
        funding_address_peer = peer.funding_address

        ret = self._htlc_close_curr_tx(
            peer=peer,
            channel_id=channel_id,
            last_rsmc_temp_address_private_key=address_self.rmsc_address.private_key,
            last_htlc_temp_address_private_key=address_self.htlc_address.private_key,
            last_htlc_temp_address_for_htnx_private_key=address_self.htlc_ht1a_address.private_key,
            curr_temp_address_index=address_self.temp_address.index,
            curr_temp_address_pub_key=address_self.temp_address.public_key,
        )

        ret_alice_sign_rsmc, ret_bob_sign_rsmc = self._alice_sign_rsmc_for_c4a(
            peer=peer,
            channel_id=channel_id,
            rsmc_partial_signed_hex=self._sign_inputs_p2skh(
                ret["c4a_rsmc_raw_data"],
                funding_address_self.private_key,
            ),
            counterparty_partial_signed_hex=self._sign_inputs_p2skh(
                ret["c4a_counterparty_raw_data"],
                funding_address_self.private_key,
            ),
        )

        resp_bob_sign_close_htlc = peer._bob_sign_close_htlc(
            peer=self,
            msg_hash=ret_bob_sign_rsmc["msg_hash"],
            c4a_rsmc_complete_signed_hex=peer._sign_inputs_p2skh(
                ret_bob_sign_rsmc["c4a_rsmc_partial_signed_data"],
                funding_address_peer.private_key,
            ),
            c4a_counterparty_complete_signed_hex=peer._sign_inputs_p2skh(
                ret_bob_sign_rsmc["c4a_counterparty_partial_signed_data"],
                funding_address_peer.private_key,
            ),
            curr_temp_address_index=address_peer.temp_address.index,
            curr_temp_address_pub_key=address_peer.temp_address.public_key,
            last_rsmc_temp_address_private_key=address_peer.rmsc_address.private_key,
            last_htlc_temp_address_for_htnx_private_key=address_peer.he_address.private_key,
            last_htlc_temp_address_private_key=address_peer.htlc_address.private_key,
        )

        resp_bob_sign_rsmc, resp_bob_sign_rsmc_alice = peer._bob_signed_rsmc_for_c4b(
            peer=self,
            channel_id=resp_bob_sign_close_htlc["channel_id"],
            c4a_rd_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_sign_close_htlc["c4a_rd_raw_data"],
                funding_address_peer.private_key,
            ),
            c4a_br_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_sign_close_htlc["c4a_br_raw_data"],
                funding_address_peer.private_key,
            ),
            c4a_br_id=resp_bob_sign_close_htlc["c4a_br_raw_data"]["br_id"],
            c4b_rsmc_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_sign_close_htlc["c4b_rsmc_raw_data"],
                funding_address_peer.private_key,
            ),
            c4b_counterparty_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_sign_close_htlc["c4b_counterparty_raw_data"],
                funding_address_peer.private_key,
            ),
        )

        resp_on_alice_signed_cxb = self._on_alice_signed_cxb(
            peer,
            channel_id=resp_bob_sign_rsmc_alice["channel_id"],
            c4a_rd_complete_signed_hex=self._sign_inputs_p2skh(
                resp_bob_sign_rsmc_alice["c4a_rd_partial_signed_data"],
                address_self.temp_address.private_key,
            ),
            c4b_rsmc_complete_signed_hex=self._sign_inputs_p2skh(
                resp_bob_sign_rsmc_alice["c4b_rsmc_partial_signed_data"],
                funding_address_self.private_key,
            ),
            c4b_counterparty_complete_signed_hex=self._sign_inputs_p2skh(
                resp_bob_sign_rsmc_alice["c4b_counterparty_partial_signed_data"],
                funding_address_self.private_key,
            ),
        )

        resp_alice_c4b, resp_bob_c4b = self._on_alice_signed_c4b(
            peer=peer,
            channel_id=resp_on_alice_signed_cxb["channel_id"],
            c4b_rd_partial_signed_hex=self._sign_inputs_p2skh(
                resp_on_alice_signed_cxb["c4b_rd_raw_data"],
                funding_address_self.private_key,
            ),
            c4b_br_partial_signed_hex=self._sign_inputs_p2skh(
                resp_on_alice_signed_cxb["c4b_br_raw_data"],
                funding_address_self.private_key,
            ),
            c4b_br_id=resp_on_alice_signed_cxb["c4b_br_raw_data"]["br_id"],
        )

        alice_resp_cxb = peer._bob_signed_cxb_sub_tx(
            self,
            channel_id=resp_bob_c4b["channel_id"],
            c4b_rd_complete_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_c4b["c4b_rd_partial_signed_data"],
                address_peer.temp_address.private_key,
            ),
        )

        return alice_resp_cxb

    def bob_signed_c3b_sub_tx(
        self,
        peer: "OmniBolt",
        channel_id: str,
        curr_htlc_temp_address_for_he_index: int,
        curr_htlc_temp_address_for_he_pub_key: str,
        c3a_htlc_htrd_complete_signed_hex: str,
        c3a_htlc_htbr_partial_signed_hex: str,
        c3a_htlc_hed_partial_signed_hex: str,
        c3b_rsmc_rd_complete_signed_hex: str,
        c3b_htlc_htd_complete_signed_hex: str,
        c3b_htlc_hlock_complete_signed_hex: str,
    ):
        message = {
            "type": -100104,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "curr_htlc_temp_address_for_he_index": curr_htlc_temp_address_for_he_index,
                "curr_htlc_temp_address_for_he_pub_key": curr_htlc_temp_address_for_he_pub_key,
                "c3a_htlc_htrd_complete_signed_hex": c3a_htlc_htrd_complete_signed_hex,
                "c3a_htlc_htbr_partial_signed_hex": c3a_htlc_htbr_partial_signed_hex,
                "c3a_htlc_hed_partial_signed_hex": c3a_htlc_hed_partial_signed_hex,
                "c3b_rsmc_rd_complete_signed_hex": c3b_rsmc_rd_complete_signed_hex,
                "c3b_htlc_htd_complete_signed_hex": c3b_htlc_htd_complete_signed_hex,
                "c3b_htlc_hlock_complete_signed_hex": c3b_htlc_hlock_complete_signed_hex,
            },
        }

        return self._sent_sync_request(message)

    def alice_signed_htlc_c3b(
        self,
        channel_id: str,
        c3a_rsmc_rd_complete_signed_hex: str,
        c3a_htlc_ht_complete_signed_hex: str,
        c3a_htlc_hlock_complete_signed_hex: str,
        c3b_rsmc_complete_signed_hex: str,
        c3b_counterparty_complete_signed_hex: str,
        c3b_htlc_complete_signed_hex: str,
    ):
        message = {
            "type": -100102,
            "recipient_node_peer_id": self.node_id,
            "recipient_user_peer_id": self.user_id,
            "data": {
                "channel_id": channel_id,
                "c3a_rsmc_rd_complete_signed_hex": c3a_rsmc_rd_complete_signed_hex,
                "c3a_htlc_ht_complete_signed_hex": c3a_htlc_ht_complete_signed_hex,
                "c3a_htlc_hlock_complete_signed_hex": c3a_htlc_hlock_complete_signed_hex,
                "c3b_rsmc_complete_signed_hex": c3b_rsmc_complete_signed_hex,
                "c3b_counterparty_complete_signed_hex": c3b_counterparty_complete_signed_hex,
                "c3b_htlc_complete_signed_hex": c3b_htlc_complete_signed_hex,
            },
        }

        return self._sent_sync_request(message)

    def alice_signed_c3b_sub_tx_at_alice_side(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c3a_htlc_htrd_partial_signed_hex: str,
        c3b_rsmc_rd_partial_signed_hex: str,
        c3b_rsmc_br_partial_signed_hex: str,
        c3b_htlc_htd_partial_signed_hex: str,
        c3b_htlc_hlock_partial_signed_hex: str,
        c3b_htlc_br_partial_signed_hex: str,
    ):
        message = {
            "type": -100103,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3a_htlc_htrd_partial_signed_hex": c3a_htlc_htrd_partial_signed_hex,
                "c3b_rsmc_rd_partial_signed_hex": c3b_rsmc_rd_partial_signed_hex,
                "c3b_rsmc_br_partial_signed_hex": c3b_rsmc_br_partial_signed_hex,
                "c3b_htlc_htd_partial_signed_hex": c3b_htlc_htd_partial_signed_hex,
                "c3b_htlc_hlock_partial_signed_hex": c3b_htlc_hlock_partial_signed_hex,
                "c3b_htlc_br_partial_signed_hex": c3b_htlc_br_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def on_bob_sign_htrd(
        self,
        peer: "OmniBolt",
        channel_id: str,
        c3b_htlc_hlock_he_partial_signed_hex: str,
    ):
        message = {
            "type": -100105,
            "recipient_node_peer_id": peer.node_id,
            "recipient_user_peer_id": peer.user_id,
            "data": {
                "channel_id": channel_id,
                "c3b_htlc_hlock_he_partial_signed_hex": c3b_htlc_hlock_he_partial_signed_hex,
            },
        }

        return self._sent_sync_request(message), peer.read_message()

    def _get_inputs_or_hex(self, sign_data, private_key):
        if sign_data["inputs"]:
            return self.sign_hex_p2skh(
                sign_data["hex"],
                sign_data["inputs"],
                [
                    sign_data["pub_key_a"],
                    sign_data["pub_key_b"],
                ],
                private_key=private_key,
            )

        else:
            return sign_data["hex"]

    def _add_htlc_workflow(self, peer: "OmniBolt", routing_information: dict[str, Any]):

        memo = routing_information["memo"]
        amount = routing_information["amount"]
        h = routing_information["h"]
        min_cltv_expiry = routing_information["min_cltv_expiry"]
        routing_packet = routing_information["routing_packet"]

        self.channel_addresses[routing_packet].rmsc_address = self.generate_address()
        self.channel_addresses[routing_packet].htlc_address = self.generate_address()
        self.channel_addresses[
            routing_packet
        ].htlc_ht1a_address = self.generate_address()

        peer.channel_addresses[routing_packet].rmsc_address = self.generate_address()
        peer.channel_addresses[routing_packet].htlc_address = self.generate_address()
        peer.channel_addresses[
            routing_packet
        ].htlc_ht1a_address = self.generate_address()

        current_commit_tx_address = self.channel_addresses[routing_packet]
        current_commit_tx_address_peer = peer.channel_addresses[routing_packet]

        last_temp_address_private = self.funding_address.private_key
        last_temp_address_private_peer = peer.funding_address.private_key
        add_htlc_response = self._add_htlc(
            peer,
            amount=amount,
            h=h,
            memo=memo,
            cltv_expiry=min_cltv_expiry,
            routing_packet=routing_packet,
            curr_rsmc_temp_address_index=current_commit_tx_address.rmsc_address.index,
            curr_rsmc_temp_address_pub_key=current_commit_tx_address.rmsc_address.public_key,
            curr_htlc_temp_address_for_ht1a_pub_key=current_commit_tx_address.htlc_ht1a_address.public_key,
            curr_htlc_temp_address_for_ht1a_index=current_commit_tx_address.htlc_ht1a_address.index,
            curr_htlc_temp_address_pub_key=current_commit_tx_address.htlc_address.public_key,
            last_temp_address_private_key=current_commit_tx_address.temp_address.private_key,
        )

        resp_a, resp_b = self.alice_sign_c3a(
            peer,
            channel_id=add_htlc_response["channel_id"],
            rsmc_partial_signed_hex=self._sign_inputs_p2skh(
                add_htlc_response["c3a_rsmc_raw_data"],
                last_temp_address_private,
            ),
            counterparty_partial_signed_hex=self._get_inputs_or_hex(
                add_htlc_response["c3a_counterparty_raw_data"],
                last_temp_address_private,
            ),
            htlc_partial_signed_hex=self._sign_inputs_p2skh(
                add_htlc_response["c3a_htlc_raw_data"],
                last_temp_address_private,
            ),
        )

        sign_bob_response = peer.sign_bob_side(
            payer_commitment_tx_hash=resp_a["commitment_tx_hash"],
            c3a_complete_signed_rsmc_hex=peer._sign_inputs_p2skh(
                resp_b["c3a_rsmc_partial_signed_data"],
                last_temp_address_private_peer,
            ),
            c3a_complete_signed_counterparty_hex=self._get_inputs_or_hex(
                resp_b["c3a_counterparty_partial_signed_data"],
                last_temp_address_private,
            ),
            c3a_complete_signed_htlc_hex=peer._sign_inputs_p2skh(
                resp_b["c3a_htlc_partial_signed_data"],
                last_temp_address_private_peer,
            ),
            last_temp_address_private_key=current_commit_tx_address_peer.temp_address.private_key,
            curr_rsmc_temp_address_index=current_commit_tx_address_peer.rmsc_address.index,
            curr_rsmc_temp_address_pub_key=current_commit_tx_address_peer.rmsc_address.public_key,
            curr_htlc_temp_address_index=current_commit_tx_address_peer.htlc_address.index,
            curr_htlc_temp_address_pub_key=current_commit_tx_address_peer.htlc_address.public_key,
        )

        response_sign_c3b_bob, response_sign_c3a_alice = peer.sign_c3b_bob_side(
            self,
            channel_id=sign_bob_response["channel_id"],
            c3a_rsmc_rd_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3a_rsmc_rd_raw_data"],
                last_temp_address_private_peer,
            ),
            c3a_rsmc_br_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3a_rsmc_br_raw_data"],
                last_temp_address_private_peer,
            ),
            c3a_htlc_ht_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3a_htlc_ht_raw_data"],
                pub_key_override=[
                    sign_bob_response["c3a_htlc_ht_raw_data"]["pub_key_b"],
                    sign_bob_response["c3a_htlc_ht_raw_data"]["pub_key_a"],
                ],
                private_key=last_temp_address_private_peer,
            ),
            c3a_htlc_hlock_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3a_htlc_hlock_raw_data"],
                last_temp_address_private_peer,
            ),
            c3a_htlc_br_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3a_htlc_br_raw_data"],
                last_temp_address_private_peer,
            ),
            c3b_rsmc_partial_signed_hex=self._get_inputs_or_hex(
                sign_bob_response["c3b_rsmc_raw_data"],
                last_temp_address_private,
            ),
            c3b_counterparty_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3b_counterparty_raw_data"],
                last_temp_address_private_peer,
            ),
            c3b_htlc_partial_signed_hex=peer._sign_inputs_p2skh(
                sign_bob_response["c3b_htlc_raw_data"],
                last_temp_address_private_peer,
            ),
        )

        alice_signed_htlc_c3b_resp = self.alice_signed_htlc_c3b(
            channel_id=response_sign_c3a_alice["channel_id"],
            c3a_rsmc_rd_complete_signed_hex=self._sign_inputs_p2skh(
                response_sign_c3a_alice["c3a_rsmc_rd_partial_signed_data"],
                current_commit_tx_address.rmsc_address.private_key,
            ),
            c3a_htlc_ht_complete_signed_hex=self._sign_inputs_p2skh(
                response_sign_c3a_alice["c3a_htlc_ht_partial_signed_data"],
                current_commit_tx_address.htlc_address.private_key,
            ),
            c3a_htlc_hlock_complete_signed_hex=self._sign_inputs_p2skh(
                response_sign_c3a_alice["c3a_htlc_hlock_partial_signed_data"],
                current_commit_tx_address.htlc_address.private_key,
            ),
            c3b_rsmc_complete_signed_hex=self._get_inputs_or_hex(
                response_sign_c3a_alice["c3b_rsmc_partial_signed_data"],
                last_temp_address_private,
            ),
            c3b_counterparty_complete_signed_hex=self._sign_inputs_p2skh(
                response_sign_c3a_alice["c3b_counterparty_partial_signed_data"],
                last_temp_address_private,
            ),
            c3b_htlc_complete_signed_hex=self._sign_inputs_p2skh(
                response_sign_c3a_alice["c3b_htlc_partial_signed_data"],
                last_temp_address_private,
            ),
        )

        (
            alice_signed_c3b_sub_tx_at_alice_side_resp,
            bob_response,
        ) = self.alice_signed_c3b_sub_tx_at_alice_side(
            peer,
            channel_id=alice_signed_htlc_c3b_resp["channel_id"],
            c3a_htlc_htrd_partial_signed_hex=self._sign_inputs_p2skh(
                alice_signed_htlc_c3b_resp["c3a_htlc_htrd_raw_data"],
                current_commit_tx_address.htlc_ht1a_address.private_key,
            ),
            c3b_rsmc_rd_partial_signed_hex=self._get_inputs_or_hex(
                alice_signed_htlc_c3b_resp["c3b_rsmc_rd_raw_data"],
                last_temp_address_private,
            ),
            c3b_rsmc_br_partial_signed_hex=self._get_inputs_or_hex(
                alice_signed_htlc_c3b_resp["c3b_rsmc_br_raw_data"],
                last_temp_address_private,
            ),
            c3b_htlc_htd_partial_signed_hex=self._sign_inputs_p2skh(
                alice_signed_htlc_c3b_resp["c3b_htlc_htd_raw_data"],
                last_temp_address_private,
            ),
            c3b_htlc_hlock_partial_signed_hex=self._sign_inputs_p2skh(
                alice_signed_htlc_c3b_resp["c3b_htlc_hlock_raw_data"],
                last_temp_address_private,
            ),
            c3b_htlc_br_partial_signed_hex=self._sign_inputs_p2skh(
                alice_signed_htlc_c3b_resp["c3b_htlc_br_raw_data"],
                last_temp_address_private,
            ),
        )

        new_address_he = peer.generate_address()
        self.channel_addresses[routing_packet].he_address = new_address_he
        peer.channel_addresses[routing_packet].he_address = new_address_he

        resp_bob_sign_c3a_sub = peer.bob_signed_c3b_sub_tx(
            self,
            channel_id=bob_response["channel_id"],
            curr_htlc_temp_address_for_he_index=new_address_he.index,
            curr_htlc_temp_address_for_he_pub_key=new_address_he.public_key,
            c3a_htlc_htrd_complete_signed_hex=peer._sign_inputs_p2skh(
                bob_response["c3a_htlc_htrd_partial_data"],
                last_temp_address_private_peer,
            ),
            c3a_htlc_htbr_partial_signed_hex=peer._sign_inputs_p2skh(
                bob_response["c3a_htlc_htrd_partial_data"],
                last_temp_address_private_peer,
            ),
            c3a_htlc_hed_partial_signed_hex=peer._sign_inputs_p2skh(
                bob_response["c3a_htlc_hed_raw_data"],
                last_temp_address_private_peer,
            ),
            c3b_rsmc_rd_complete_signed_hex=self._get_inputs_or_hex(
                bob_response["c3b_rsmc_rd_partial_data"],
                current_commit_tx_address_peer.rmsc_address.private_key,
            ),
            c3b_htlc_htd_complete_signed_hex=peer._sign_inputs_p2skh(
                bob_response["c3b_htlc_htd_partial_data"],
                current_commit_tx_address_peer.htlc_address.private_key,
            ),
            c3b_htlc_hlock_complete_signed_hex=peer._sign_inputs_p2skh(
                bob_response["c3b_htlc_hlock_partial_data"],
                current_commit_tx_address_peer.htlc_address.private_key,
            ),
        )

        on_bob_sign_htrd, alice_response_bob_sign_htrd = peer.on_bob_sign_htrd(
            self,
            channel_id=resp_bob_sign_c3a_sub["channel_id"],
            c3b_htlc_hlock_he_partial_signed_hex=peer._sign_inputs_p2skh(
                resp_bob_sign_c3a_sub["c3b_htlc_hlock_he_raw_data"],
                last_temp_address_private_peer,
            ),
        )

        return on_bob_sign_htrd, alice_response_bob_sign_htrd

    def pay_invoice(self, invoice: str, peer: "OmniBolt"):
        routing_information = self._get_routing_information(invoice, peer)

        on_bob_sign_htrd, alice_response_bob_sign_htrd = self._add_htlc_workflow(
            peer, routing_information
        )
        channel_id = on_bob_sign_htrd["channel_id"]

        peer.backwards_workflow(
            self,
            channel_id=channel_id,
            r=peer.h_address.private_key,
        )

        close_resp = self.close_workflow(
            peer,
            channel_id=channel_id,
        )

        return close_resp

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

        temp_address_rsmc = self.generate_address()
        temp_address_htlc_ht1a = self.generate_address()
        temp_address_htlc = self.generate_address()

        temp_address_rsmc_bob = self.generate_address()
        temp_address_htlc_bob = self.generate_address()
        temp_address_htlc_ht1a_bob = self.generate_address()

        open_channel_resp, open_channel_resp_bob = self._open_channel(
            peer_node=peer, funding_public_key=temp_address_rsmc.public_key
        )

        temp_channel_id = open_channel_resp["temporary_channel_id"]

        accept_channel_response, funding_accept_reply = peer.accept_channel(
            peer_node=self,
            channel_id=temp_channel_id,
            funding_public_key=temp_address_rsmc_bob.public_key,
        )
        channel_address = funding_accept_reply["channel_address"]
        omnicore_connection.send_bitcoin(temp_address_rsmc.address, 1000000000)
        omnicore_connection.mine_bitcoin(50, self.wallet_address.address)

        grant_amount = "9999999999.00000000"
        omnicore_connection.send_omnicore_token(
            funding_btc,
            temp_address_rsmc.address,
            omnicore_item,
            grant_amount,
        )
        funding_count: Final = 3
        for _ in range(funding_count):
            funding_btc_response = self.funding_btc(
                from_address=temp_address_rsmc.address,
                from_address_private=temp_address_rsmc.private_key,
                to_address=channel_address,
            )

            signed_tx_hex = self.sign_hex(
                funding_btc_response["hex"],
                funding_btc_response["inputs"],
                temp_address_rsmc.private_key,
            )

            omnicore_connection.mine_bitcoin(2, self.wallet_address.address)

            funding_created_response = self.btc_funding_created(
                peer_node=peer,
                channel_id=temp_channel_id,
                funding_tx_hex=signed_tx_hex,
            )

            funding_created_response_hex = self.sign_hex_p2skh(
                funding_created_response["hex"],
                funding_created_response["inputs"],
                [
                    funding_created_response["pub_key_a"],
                    funding_created_response["pub_key_b"],
                ],
                temp_address_rsmc.private_key,
            )

            response_sign_mine, response_sign_mine_bob = self.sign_mining(
                peer, funding_created_response_hex
            )

            response_sign_mine_signed = peer.sign_hex(
                response_sign_mine["sign_data"]["hex"],
                response_sign_mine["sign_data"]["inputs"],
                temp_address_rsmc_bob.private_key,
            )

            btc_funding_created_response = self.btc_funding_created(
                peer_node=peer,
                channel_id=temp_channel_id,
                funding_tx_hex=signed_tx_hex,
            )

            peer.read_message()

            peer.btc_funding_singed(
                peer_node=self,
                channel_id=temp_channel_id,
                funding_tx_id=btc_funding_created_response["funding_txid"],
                signed_miner_redeem_transaction_hex=response_sign_mine_signed,
            )

        funding_asset_response = self.funding_assert(
            temp_address_rsmc.address,
            channel_address,
            channel_size,
            property_id,
        )

        channel_addresses = ChannelAddresses(
            htlc_address=temp_address_htlc,
            htlc_ht1a_address=temp_address_htlc_ht1a,
            rmsc_address=temp_address_rsmc,
            he_address=self.generate_address(),
            temp_address=temp_address_rsmc,
        )

        self.funding_address = temp_address_rsmc
        peer.funding_address = temp_address_rsmc_bob

        channel_addresses_bob = ChannelAddresses(
            htlc_address=temp_address_htlc_bob,
            htlc_ht1a_address=temp_address_htlc_ht1a_bob,
            rmsc_address=temp_address_rsmc_bob,
            he_address=peer.generate_address(),
            temp_address=temp_address_rsmc_bob,
        )

        funding_created_response = self.asset_funding_created(
            peer_node=peer,
            channel_id=temp_channel_id,
            funding_tx_hex=self.sign_hex(
                funding_asset_response["hex"],
                funding_asset_response["inputs"],
                temp_address_rsmc.private_key,
            ),
            temp_address_public_key=temp_address_rsmc.public_key,
            temp_address_index=temp_address_rsmc.index,
        )

        response_sign_c1a, response_bob = self.sign_hex_c1a(
            peer,
            self.sign_hex_p2skh(
                funding_created_response["hex"],
                funding_created_response["inputs"],
                [
                    funding_created_response["pub_key_a"],
                    funding_created_response["pub_key_b"],
                ],
                temp_address_rsmc.private_key,
            ),
        )

        response_bob = peer.asset_funding_signed(
            peer_node=self,
            temporary_channel_id=temp_channel_id,
            signed_funding_assert_hex=peer.sign_hex_p2skh(
                response_sign_c1a["sign_data"]["hex"],
                response_sign_c1a["sign_data"]["inputs"],
                [
                    response_sign_c1a["sign_data"]["pub_key_a"],
                    response_sign_c1a["sign_data"]["pub_key_b"],
                ],
                temp_address_rsmc_bob.private_key,
            ),
        )
        alice_sign_data_br = response_bob["alice_br_sign_data"]
        alice_sign_data_rd = response_bob["alice_rd_sign_data"]

        bob_sign_response, bob_sign_response_alice = peer.signed_rd_and_br(
            peer_node=self,
            temporary_channel_id=temp_channel_id,
            br_id=alice_sign_data_br["br_id"],
            br_signed_hex=peer.sign_hex_p2skh(
                alice_sign_data_br["hex"],
                alice_sign_data_br["inputs"],
                [alice_sign_data_br["pub_key_a"], alice_sign_data_br["pub_key_b"]],
                temp_address_rsmc_bob.private_key,
            ),
            rd_signed_hex=peer.sign_hex_p2skh(
                alice_sign_data_rd["hex"],
                alice_sign_data_rd["inputs"],
                [alice_sign_data_rd["pub_key_a"], alice_sign_data_rd["pub_key_b"]],
                temp_address_rsmc_bob.private_key,
            ),
        )
        channel_id = bob_sign_response["channel_id"]

        self.alice_signed_rd_of_asset_funding(
            channel_id=channel_id, rd_signed_hex=bob_sign_response_alice["hex"]
        )

        self.channel_addresses[channel_id] = channel_addresses
        peer.channel_addresses[channel_id] = channel_addresses_bob

        return channel_id
