'use strict';

var StepRecorder = function(){
    
};

StepRecorder.prototype = {
    record: function(key,steps){
        LocalStorage.set(String(key), String(steps));
        _log.info("-----------------------------" +
                "the save ok" +
                "-------------------------------------");
        return ;
    },
    read: function(key){
        var originalSteps = LocalStorage.get(String(key));
        _log.info("-----------------------------" +
                "the return value is ok" +
                "-------------------------------------"
                +originalSteps);
            event.trigger("sc_dappley_block_data", originalSteps)

        return;
    },
    destory: function(addr, steps){
        var address = Sender.getAddress()
        if (addr == address){
            Blockchain.deleteContract()
        }
    },
    depCheck:function () {
        _log.info("\n\n deployed successfully \n\n");
    },
    dapp_schedule: function(){}
};

module.exports = new StepRecorder;