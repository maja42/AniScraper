'use strict';

angular.module('AniScraper')
.service('SocketService', function($log, $rootScope, $timeout){
    var messageResponseTimeout = 10000;     // number of milliseconds until waiting for a message response is aborted with a timeout.
    var socket = null;
    var connected = false;
    var correlationIdSequence = 0;

    var futures = {};       // map containing all callback functions for message responses.

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

        if(messageObject.responseFor && messageObject.responseFor > 0) {
            // This is a response message
            messageResponseReceived(messageObject.responseFor, messageObject.messageType, messageObject.message);
        } else {
            $rootScope.$emit("websocket-message-event", messageObject.messageType, messageObject.message);
        }
    }

    this.send = function(messageType, message, success, error) {
        if(!this.isConnected()) {
            $log.error("Unable to send message: Not connected. \nMessage type:", messageType, "\nMessage:", message);
            return;
        }
        var expectsResponse = success || error;

        var messageObject = {
            "messageType": messageType,
            "message": message
        };

        if(expectsResponse) {
            // This message is expecting a response
            var corId = nextCorrelationId();
            messageObject.answerAt = corId;
            
            futures[corId] = {
                success: success,
                error: error,
                timer: $timeout(function() {
                    messageResponseAbortion(corId, "Timeout reached.");
                }, messageResponseTimeout),
            };
        }
        socket.send(JSON.stringify(messageObject));
    };

    function nextCorrelationId() {
        ++correlationIdSequence;
        return correlationIdSequence;
    }

    // Executed after a message response has been received
    function messageResponseReceived(corId, messageType, message) {
        var future = futures[corId];
        if(!future) {
            $log.error("Received a message response for unknown correlationId " + corId + ". Maybe the original message already timed out?");
            return;
        }
        futures[corId] = undefined;

        $timeout.cancel(future.timer);
        if(future.success) {
            future.success(messageType, message);
        }
    }

    // Executed if we don't expect a response for the given message anymore. For example after timeout has been reached (the server didn't respond to a message which expected one in time)
    function messageResponseAbortion(corId, errorMessage) {
        var future = futures[corId];
        futures[corId] = undefined;
        if(future.error) {
            future.error(errorMessage);
        }
    }

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


    function destroy() {
        // Cancel all running timers and inform waiting listeners that they won't get a response anytime soon
        for(corId in futures) {
            $timeout.cancel(futures[corId].timer);
            messageResponseAbortion(corId, "SocketService destroyed");
        }
    }

    $rootScope.$on("$destroy", function(event) {
       destroy();
    });
});
