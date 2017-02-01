'use strict';

angular.module('AniScraper')
    .directive('subMenu', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  label, items  */
            },
            controller: function ($scope) {
                $scope.toggleButton = {
                    icon: "angle-right",
                    label: $scope.properties.label,
                    onClick: function () {
                        $scope.collapse = !$scope.collapse;
                        if ($scope.collapse) {
                            $scope.toggleButton.icon = "angle-right";
                        } else {
                            $scope.toggleButton.icon = "angle-down";
                        }

                    }
                };

                $scope.collapse = true;
            },

            //todo: dieser hässliche workaround hier gehört weg.
            //Lösung: in der sidebar-direktive (sidebar.js) nicht mit einem switch alles rendern, sondern abhängig von der value das richtige template zurückgeben
            template: '<div class="sub-menu">' +
            '    <action-button properties="toggleButton"></action-button>' +
                //TODO: replace the below ul-list with the sidebar-list directive (difficult, because it created recursive digest cicle. Stupid angular design.
                //'    <sidebar-list class="nav nav-second-level" collapse="collapse" items="properties.items"></sidebar-list>' +
            '    <ul class="nav nav-second-level" collapse="collapse">' +
            '        <li ng-repeat="item in properties.items" ng-switch="item.directive">' +
            '            <redirect-button ng-switch-when="redirectButton" properties="item"></redirect-button>' +
            '            <action-button   ng-switch-when="actionButton" properties="item"></action-button>' +
            '            <search-input    ng-switch-when="searchInput" properties="item"></search-input>' +
            '            <checkbox-input  ng-switch-when="checkboxInput" properties="item"></checkbox-input>' +
            '            <dropdown-input  ng-switch-when="dropdownInput" properties="item"></dropdown-input>' +
            '            <div ng-switch-default>{{item}}</div>' +
            '        </li>' +
            '    </ul>' +
            '</div>'
        }
    })
    
    //THIS directive is also used in the header (header pages)
    .directive('redirectButton', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /* icon, iconSize, label, state, stateParam */
            },
            link: function ($scope) {
                $scope.iconSize = "fa-" + ($scope.properties.iconSize || "fw");
            },
            template: '<a class="button redirect-button" ui-sref="{{properties.state}}({{properties.stateParam}})" ui-sref-active="active">' +
            '    <i class="fa fa-{{properties.icon}} {{iconSize}}"></i>' +
            '    <span>{{properties.label}}</span>' +
            '</a>'
        }
    })

    .directive('actionButton', function () {
        function linker($scope, $element, $attrs) {
            $scope.properties.iconSize = $scope.properties.iconSize || "fw";
            if (!$scope.properties.hasOwnProperty("enabled")) {
                $scope.properties.enabled = true;
            }
            $scope.onButtonClick = function () {
                if ($scope.properties.enabled) {
                    $scope.properties.onClick();
                }
            };
        }

        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  enabled, icon, label, onClick  */
            },
            link: linker,
            template: '<a class="button action-button" ng-class="{disabled: !properties.enabled}" ng-click="onButtonClick()">' +
            '    <i class="fa fa-{{properties.icon}} fa-{{properties.iconSize}}"></i>' +
            '    <span>{{properties.label}}</span>' +
            '</a>'
        }
    })

    .directive('searchInput', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  placeholder, input, onChange, search  */
            },
            controller: function ($scope) {
                $scope.searchInput = "";
            },
            template: '<div class="input-group input search-input">' +
            '    <input type="text" class="form-control" placeholder="{{properties.placeholder}}" ng-model="properties.input" ng-change="properties.onChange(properties.input)">' +
            '    <span class="input-group-btn">' +
            '    <button class="btn btn-default" type="button" ng-click="properties.search(properties.input)">' +
            '        <i class="fa fa-search"></i>' +
            '    </button>' +
            '    </span>' +
            '</div>'
        }
    })

    .directive('sidebarTagsInput', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  placeholder, minLength, maxLength, onChange, getSuggestion, input */
            },
            template: '<div class="input-group input search-input">' +
            '   <tags-input class="form-control customer-tag-input outline" style="height: inherit" ' +
            '       on-tag-added="properties.onChange(properties.input)" on-tag-removed="properties.onChange(properties.input)" ' +
            '       ng-model="properties.input" placeholder="{{properties.placeholder}}" ' +
            '       min-length="{{properties.minLength}}" max-length="{{properties.maxLength}}">' +
            '       <auto-complete source="properties.getSuggestion($query)"' +
            '           highlight-matched-text="false" min-length="1" load-on-focus="false" load-on-empty="false"></auto-complete>' +
            '   </tags-input>' +
            '</div>'
        }
    })

    .directive('checkboxInput', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  icon, label, model, onChange  */
            },
            template: '<div class="input checkbox-input">' +
            '    <label>' +
            '        <i class="fa fa-{{properties.icon}} fa-fw"></i>' +
            '        <span>{{properties.label}}</span>' +
            '        <input type="checkbox" ng-model="properties.model" ng-change="properties.onChange(properties.model)">' +
            '    </label>' +
            '</div>'
        }
    })

    .directive('dropdownInput', function () {
        return {
            restrict: 'E',
            replace: true,
            scope: {
                properties: "="
                /*  icon, label, options, model, onChange  */
            },
            template: '<div class="input dropdown-input">' +
            '    <label>' +
            '        <i class="fa fa-{{properties.icon}} fa-fw"></i>' +
            '        <span>{{properties.label}}</span>' +
            '        <select ng-model="properties.model" ng-change="properties.onChange(properties.model)"  ng-options="key as value for (key, value) in properties.options">' +
            '        </select>' +
            '    </label>' +
            '</div>'
        }
    })

;
