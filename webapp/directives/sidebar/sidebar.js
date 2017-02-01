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
                SocketService.subscribe($scope, "echo", function(){
                    console.log("ASDF");
                });
            }
        }
    }])
    .directive('sidebarList', function() {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                itemList: "=items"
            },
            template:   '<ul>' +
            '   <li ng-repeat="item in itemList" ng-class="{\'noSeparation\': item.noSeparation}" ng-switch="item.directive">' +
            '       <sub-menu           ng-switch-when="subMenu" properties="item"></sub-menu>' +
            '       <redirect-button    ng-switch-when="redirectButton" properties="item"></redirect-button>' +
            '       <action-button      ng-switch-when="actionButton" properties="item"></action-button>' +
            '       <search-input       ng-switch-when="searchInput" properties="item"></search-input>' +
            '       <sidebar-tags-input ng-switch-when="tagsInput" properties="item"></sidebar-tags-input>' +
            '       <checkbox-input     ng-switch-when="checkboxInput" properties="item"></checkbox-input>' +
            '       <dropdown-input     ng-switch-when="dropdownInput" properties="item"></dropdown-input>' +
            '       <sidebar-billing-list ng-switch-when="billingList" orders="item.items"></sidebar-billing-list>' +
            '       <div ng-switch-default>{{item}}</div>' +
            '   </li>' +
            '</ul>'
        }
    });
