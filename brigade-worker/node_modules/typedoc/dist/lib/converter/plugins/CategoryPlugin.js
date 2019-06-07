"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var CategoryPlugin_1;
const index_1 = require("../../models/reflections/index");
const ReflectionCategory_1 = require("../../models/ReflectionCategory");
const components_1 = require("../components");
const converter_1 = require("../converter");
const GroupPlugin_1 = require("./GroupPlugin");
let CategoryPlugin = CategoryPlugin_1 = class CategoryPlugin extends components_1.ConverterComponent {
    initialize() {
        this.listenTo(this.owner, {
            [converter_1.Converter.EVENT_RESOLVE]: this.onResolve,
            [converter_1.Converter.EVENT_RESOLVE_END]: this.onEndResolve
        });
    }
    onResolve(context, reflection) {
        if (reflection instanceof index_1.ContainerReflection) {
            if (reflection.children && reflection.children.length > 0) {
                reflection.children.sort(GroupPlugin_1.GroupPlugin.sortCallback);
                reflection.categories = CategoryPlugin_1.getReflectionCategories(reflection.children);
            }
            if (reflection.categories && reflection.categories.length > 1) {
                reflection.categories.sort(CategoryPlugin_1.sortCatCallback);
            }
        }
    }
    onEndResolve(context) {
        function walkDirectory(directory) {
            directory.categories = CategoryPlugin_1.getReflectionCategories(directory.getAllReflections());
            for (let key in directory.directories) {
                if (!directory.directories.hasOwnProperty(key)) {
                    continue;
                }
                walkDirectory(directory.directories[key]);
            }
        }
        const project = context.project;
        if (project.children && project.children.length > 0) {
            project.children.sort(GroupPlugin_1.GroupPlugin.sortCallback);
            project.categories = CategoryPlugin_1.getReflectionCategories(project.children);
        }
        if (project.categories && project.categories.length > 1) {
            project.categories.sort(CategoryPlugin_1.sortCatCallback);
        }
        walkDirectory(project.directory);
        project.files.forEach((file) => {
            file.categories = CategoryPlugin_1.getReflectionCategories(file.reflections);
        });
    }
    static getReflectionCategories(reflections) {
        const categories = [];
        reflections.forEach((child) => {
            const childCat = CategoryPlugin_1.getCategory(child);
            if (childCat === '') {
                return;
            }
            for (let i = 0; i < categories.length; i++) {
                const category = categories[i];
                if (category.title !== childCat) {
                    continue;
                }
                category.children.push(child);
                return;
            }
            const category = new ReflectionCategory_1.ReflectionCategory(childCat);
            category.children.push(child);
            categories.push(category);
        });
        return categories;
    }
    static getCategory(reflection) {
        if (reflection.comment) {
            const tags = reflection.comment.tags;
            if (tags) {
                for (let i = 0; i < tags.length; i++) {
                    if (tags[i].tagName === 'category') {
                        let tag = tags[i].text;
                        return (tag.charAt(0).toUpperCase() + tag.slice(1).toLowerCase()).trim();
                    }
                }
            }
        }
        return '';
    }
    static sortCallback(a, b) {
        return a.name > b.name ? 1 : -1;
    }
    static sortCatCallback(a, b) {
        const aWeight = CategoryPlugin_1.WEIGHTS.indexOf(a.title);
        const bWeight = CategoryPlugin_1.WEIGHTS.indexOf(b.title);
        if (aWeight < 0 && bWeight < 0) {
            return a.title > b.title ? 1 : -1;
        }
        if (aWeight < 0) {
            return 1;
        }
        if (bWeight < 0) {
            return -1;
        }
        return aWeight - bWeight;
    }
};
CategoryPlugin.WEIGHTS = [];
CategoryPlugin = CategoryPlugin_1 = __decorate([
    components_1.Component({ name: 'category' })
], CategoryPlugin);
exports.CategoryPlugin = CategoryPlugin;
//# sourceMappingURL=CategoryPlugin.js.map