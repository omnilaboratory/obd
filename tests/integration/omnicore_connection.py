import os

from omnicore import rpc as bitcoin_rpc


class OmnicoreConnection:
    def __init__(self):
        self._bitcoin_connection: bitcoin_rpc.Proxy = bitcoin_rpc.Proxy(
            service_port=18330,
            btc_conf_file=os.environ.get("BITCOIN_CONF"),
        )

    def omni_sendgrant(self, address, omnicore_item, amount):

        self._bitcoin_connection.omni_sendgrant(
            address,
            address,
            omnicore_item["propertyid"],
            amount,
        )

    def send_omnicore_token(self, from_address, to_address, omnicore_item, amount):

        self._bitcoin_connection.omni_send(
            from_address, to_address, omnicore_item["propertyid"], str(amount)
        )

    def generate_bitcoin_address(self, account) -> str:
        return str(self._bitcoin_connection.getnewaddress(account))

    def get_private_key(self, address):
        return str(self._bitcoin_connection.dumpprivkey(address))

    def _generate_omni_currency(self, funder_address):
        """

        omnicore-cli -regtest sendtoaddress ${new_address_2} 50
        omnicore-cli -regtest omni_sendissuancemanaged ${new_address_2} 2 1 0 "Companies" "Bitcoin Mining" "Quantum Miner" "" ""
        omnicore-cli -regtest generatetoaddress 5 ${new_address}
        omnicore-cli -regtest omni_listproperties

        """
        return self._bitcoin_connection.omni_sendissuancemanaged(
            funder_address, 2, 2, 0, "Companies", "Bitcoin Mining", "Qunatum", "a", "a"
        )

    def list_omnicore_properties(self):
        return self._bitcoin_connection.omni_listproperties()

    def generate_omni_token(self, address_master_funder, address_miner):
        generate_omni_currency_response = self._generate_omni_currency(
            address_master_funder
        )

        print("Generate omnicore_currency_response: ", generate_omni_currency_response)
        self.mine_bitcoin(20, address_miner)

        omni_core_properties = self.list_omnicore_properties()
        print("Omnicore properties=", omni_core_properties)

        created_omnicore_item = None
        for properties in omni_core_properties:
            if generate_omni_currency_response == properties["creationtxid"]:
                created_omnicore_item = properties
                break

        if created_omnicore_item is None:
            assert False

        return created_omnicore_item

    def send_bitcoin(self, receiver: str, amount: int):
        return self._bitcoin_connection.sendtoaddress(receiver, amount)

    def mine_bitcoin(self, block_count: int, target_address: str):
        self._bitcoin_connection.generatetoaddress(block_count, target_address)
