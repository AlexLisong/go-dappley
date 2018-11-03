'use strict';

var Blockchain = function () {
    Object.defineProperty(this, "nativeBlockchain", {
        configurable: false,
        enumerable: false,
        get: function(){
            return _native_blockchain;
        }
    });
};

Blockchain.prototype = {
    verifyAddress: function (address) {
        return this.nativeBlockchain.verifyAddress(address);
    }
};
var Blockchain = new Blockchain();
//module.exports = new Blockchain();