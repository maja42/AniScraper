'use strict';

angular.module('AniScraper')
.directive('socketMonitor', ["$log", "SocketService", function($log, SocketService) {
    return {
        templateUrl:'directives/socket-monitor/socket-monitor.html',
        restrict: 'E',
        replace: true,
        scope: {
            websocketUrl: "@"
        },
        controller: function ($scope) {
           
            SocketService.subscribeToMeta($scope, "connected", function(event) {
                $log.info("Connected to websocket");
                var message = "I am alive!";
                $log.debug("Sending echo message: " + message);


                SocketService.send("echo", message, function(messageType, incomingMessage) {
                    if(messageType == "echo-reply" && message == incomingMessage) {
                        $log.info("Echo-Channel works. The server responded correctly.");
                    }else {
                        $log.error("Echo-Channel is broken. The server responded with a different message. Type=", messageType, "Message=", incomingMessage);
                    }
                }, function(errorMessage) {
                    $log.error("Echo-Channel is broken.", errorMessage);
                });
            });



            SocketService.subscribeToMeta($scope, "error", function(event) {
                $log.error("Websocket error", event);
            });

            SocketService.subscribeToMeta($scope, "disconnected", function(event) {
                $log.info("Disconnected from websocket (" + event.code + ")", event.message);
                // TODO: reconnect
            });

            SocketService.subscribe($scope, "echo", function(messageType, message) {
                $log.info("Replying to echo message: ", message);
                SocketService.send("echo-reply", message);
            });

            SocketService.connect($scope.websocketUrl);
        }
    };
}]);



