'use strict';



angular.module('AniScraper')
.config(['$stateProvider', '$urlRouterProvider', function ($stateProvider, $urlRouterProvider) {

    $urlRouterProvider.when('', '/content');
    $urlRouterProvider.otherwise('/content');


    /**
     * Returns a resolve function that utilises lazyLoad to load additional resources.
     * @param files         A single file (string) or a list of files (array of strings) that should be lazy loaded
     * @returns {Function}  Resolve function that can be used for routing
     */
    function resolveFiles(files) {
        if(typeof files === "string") {
            files = [files];
        }
        return function ($ocLazyLoad) {
            return $ocLazyLoad.load({
                name: 'sapling',
                files: files
            });
        }
    }

    $stateProvider
        .state('content', {
            url: '/content',
            controller: 'contentCtrl',
            templateUrl: 'views/content/content.html',
            // resolve: {
            //     requires: resolveFiles('views/content/homeController.js')
            // }
        })
    ;
}]);


