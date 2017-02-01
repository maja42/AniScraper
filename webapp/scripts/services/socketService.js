'use strict';

angular.module('AniScraper')
.service('SocketService', function($log, $rootScope){
    var socket = null;
    var connected = false;

    this.connect = function(url) {
        if(this.isConnected()) {
            $log.error("Already connected. Disconnect WebSocket before initiating a new connection.")
            return;
        }
        $log.info("Connecting to websocket at " + url)
        socket = new WebSocket(url);

        socket.addEventListener('open', onSocketOpen);
        socket.addEventListener('message', onSocketMessage);
        socket.addEventListener('error', onSocketError);
        socket.addEventListener('close', onSocketClose);
    };

    this.disconnect = function() {
        if(socket == null){
            $log.warn("Not connected. Unable to disconnect.");
            return;
        }
        socket.close();
    };

    this.isConnected = function() {
        if(socket == null) {
            return false;
        }
        if(socket.readyState === 3) {
            return false;
        }
        if(!connected) {
            return false;
        }
        return true;
    };

    function onSocketOpen(event) {
        connected = true;
        $rootScope.$emit("websocket-connected", event);
    }

    function onSocketError(event) {
        $rootScope.$emit("websocket-error", event);
    }

    function onSocketClose(event) {
        connected = false;
        $rootScope.$emit("websocket-disconnected", event);
    }
    
    function onSocketMessage(event) {
        var messageObject;
        try {
            messageObject = JSON.parse(event.data);
        } catch(e) {
            $log.error("Failed to parse incomming websocket message:", e, "\nMessage:", event.data);
            return;
        }
        $log.debug("Received incomming message (" + messageObject.messageType + "):", messageObject.message)
        $rootScope.$emit("websocket-message-event", messageObject.messageType, messageObject.message);
    }

    this.send = function(messageType, message) {
        if(!this.isConnected()) {
            $log.error("Unable to send message: Not connected. \nMessage type:", messageType, "\nMessage:", message);
            return;
        }
        var messageObject = {
            "messageType": messageType,
            "message": message
        };
        socket.send(JSON.stringify(messageObject));
    };

    // Subscribe to incoming messages...
    this.subscribe = function($scope, messageTypes, callback) {
        if(!angular.isArray(messageTypes)) {
            messageTypes = [messageTypes];
        }

        var handler = $rootScope.$on("websocket-message-event", function(event, messageType, message){
            if(messageTypes.indexOf(messageType) !== -1) {  // The caller is interested in this message type
                callback(messageType, message);
            }
        });
        $scope.$on('$destroy', handler);
    };

    // Subscribe to websocket meta events like connected/error/disconnected
    this.subscribeToMeta = function($scope, eventType, callback) {
        var eventTypes = {
            "connected": "websocket-connected",
            "error": "websocket-error",
            "disconnected": "websocket-disconnected",
        };
        if(!eventTypes.hasOwnProperty(eventType)) {
            $log.error("Invalid eventType: ", eventType, "- unable to subscribe");
            return;
        }

        var handler = $rootScope.$on(eventTypes[eventType], function(event, socketEvent) {
            callback(socketEvent);
        });
        $scope.$on('$destroy', handler);
    };

});
