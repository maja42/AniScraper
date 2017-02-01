'use strict';

angular.module('AniScraper')
    .directive('sidebar', ['SocketService', function(SocketService) {
        return {
            templateUrl: 'directives/sidebar/sidebar.html',
            restrict: 'E',
            replace: true,
            scope: {
                controls: "=",
                actionButtons: "="
            },
            controller: function ($scope) {
                $scope.mediaCollection = ["A", "B", "C"];


                SocketService.subscribe($scope, "clearMediaCollection", function(messageType, message){
                    $scope.mediaCollection = [];
                });

                SocketService.subscribe($scope, "newMediaFolder", function(messageType, message){
                    $scope.mediaCollection.push("aaa" + message);

                    $scope.$apply();
                });
            }
        }
    }]);
