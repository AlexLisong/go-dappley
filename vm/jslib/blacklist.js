'use strict';

var Blacklist = function () {

};

Blacklist.prototype = {
    add: function (blackAddr) {
        return _native_blacklist.blacklist("ADD",blackAddr);
    },
    del: function (blackAddr) {
        return _native_blacklist.blacklist("DEL",blackAddr);
    },
    dapp_schedule: function () {
    }
};

module.exports = new Blacklist();