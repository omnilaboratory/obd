import time

import pytest

import omnicore
from omnibolt import OmniBolt
from runner import TestRunner

omnicore.SelectParams("regtest")

_DEFAULT_CHANNEL_SIZE = 1000


@pytest.fixture
def basic_channel() -> TestRunner:
    test_runner = TestRunner()
    test_runner.setup_basic_workflow(channel_size=_DEFAULT_CHANNEL_SIZE)
    return test_runner


def test_basic_workflow(basic_channel):
    alice = basic_channel.omni_bolt_alice
    open_channels = alice.get_all_channels()
    assert len(open_channels) == 1
    assert open_channels[0]["asset_amount"] == _DEFAULT_CHANNEL_SIZE
    assert open_channels[0]["balance_a"] == _DEFAULT_CHANNEL_SIZE
    assert len(open_channels) == 1, open_channels


def _send_channel(from_: OmniBolt, to: OmniBolt, amount: int):
    open_channels_alice = from_.get_all_channels()
    amount_from = open_channels_alice[0]["balance_a"]
    amount_to = open_channels_alice[0]["balance_b"]

    invoice_bob = to.add_invoice(from_.property_id, amount=amount)
    from_.pay_invoice(invoice_bob, to)
    time.sleep(1)
    open_channels_after_pay_invoice = from_.get_all_channels()
    assert open_channels_after_pay_invoice[0]["balance_a"] == amount_from - amount
    assert open_channels_after_pay_invoice[0]["balance_b"] == amount_to + amount


def test_send_to_bob(basic_channel):
    alice = basic_channel.omni_bolt_alice
    bob = basic_channel.omni_bolt_bob
    _send_channel(alice, bob, amount=300)
    _send_channel(bob, alice, amount=300)


def test_send_to_bob_many_transactions(basic_channel):
    alice = basic_channel.omni_bolt_alice
    bob = basic_channel.omni_bolt_bob
    _send_channel(alice, bob, amount=100)
    for _ in range(100):
        _send_channel(alice, bob, amount=300)
        _send_channel(bob, alice, amount=300)
