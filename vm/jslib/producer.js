'use strict';

var Producer = function () {

};

Producer.prototype = {
    add: function (nodeAddr) {
        return _native_producer.producer("ADD",nodeAddr);
    },
    del: function (nodeAddr) {
        return _native_producer.producer("DEL",nodeAddr);
    },
    dapp_schedule: function () {
    }
};

module.exports = new Producer();