"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const abstract_1 = require("./abstract");
const lodash_1 = require("lodash");
class ContainerReflection extends abstract_1.Reflection {
    getChildrenByKind(kind) {
        return (this.children || []).filter(child => child.kindOf(kind));
    }
    traverse(callback) {
        for (const child of lodash_1.toArray(this.children)) {
            if (callback(child, abstract_1.TraverseProperty.Children) === false) {
                return;
            }
        }
    }
    toObject() {
        const result = super.toObject();
        if (this.groups) {
            const groups = [];
            this.groups.forEach((group) => {
                groups.push(group.toObject());
            });
            result['groups'] = groups;
        }
        if (this.categories) {
            const categories = [];
            this.categories.forEach((category) => {
                categories.push(category.toObject());
            });
            if (categories.length > 0) {
                result['categories'] = categories;
            }
        }
        if (this.sources) {
            const sources = [];
            this.sources.forEach((source) => {
                sources.push({
                    fileName: source.fileName,
                    line: source.line,
                    character: source.character
                });
            });
            result['sources'] = sources;
        }
        return result;
    }
}
exports.ContainerReflection = ContainerReflection;
//# sourceMappingURL=container.js.map