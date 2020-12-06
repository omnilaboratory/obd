import pytest

import omnicore
from runner import TestRunner

omnicore.SelectParams("regtest")

_DEFAULT_CHANNEL_SIZE = 100000


@pytest.fixture(scope="session")
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
