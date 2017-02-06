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
                $scope.animeCollection = ["A", "B", "C"];


                SocketService.subscribe($scope, "clearAnimeCollection", function(messageType, message){
                    $scope.animeCollection = [];
                });

                SocketService.subscribe($scope, "newAnimeFolder", function(messageType, message){
                    $scope.animeCollection.push(message);

                    $scope.$apply();
                });
            }
        }
    }]);
