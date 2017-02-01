'use strict';

angular.module('AniScraper')
.directive('header', [function() {
    return {
        templateUrl:'directives/header/header.html',
        restrict: 'E',
        replace: true,
    };
}]);