'use strict';

angular.module('AniScraper')
.directive('content', function(){
    return {
        templateUrl:'directives/content/content.html',
        restrict: 'E',
        replace: true,
        controller: function ($scope) {
            $scope.sidebar = {};
            $scope.sidebar.controls = [
                {   directive: "actionButton",   icon: "trash",              label: "Löschen",               enabled: false, onClick: function(){ alert("Not implemented yet") }}
            ];
        }
    }
});