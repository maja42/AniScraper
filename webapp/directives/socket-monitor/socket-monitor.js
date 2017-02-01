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
                SocketService.send("echo", message);
            });

            SocketService.subscribeToMeta($scope, "error", function(event) {
                $log.error("Websocket error", event);
            });

            SocketService.subscribeToMeta($scope, "disconnected", function(event) {
                $log.info("Disconnected from websocket (" + event.code + ")", event.message);
                // TODO: reconnect
            });

            SocketService.subscribe($scope, "echo-reply", function(messageType, message) {
                $log.debug("Got echo reply message:", message);
            });

            SocketService.subscribe($scope, "echo", function(messageType, message) {
                $log.debug("Replying to echo message: ", message);
                SocketService.send("echo-reply", message);
            });

            SocketService.connect($scope.websocketUrl);
        }
    };
}]);



