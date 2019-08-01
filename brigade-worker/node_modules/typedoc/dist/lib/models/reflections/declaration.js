"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const abstract_1 = require("./abstract");
const index_1 = require("../types/index");
const container_1 = require("./container");
const lodash_1 = require("lodash");
class DeclarationReflection extends container_1.ContainerReflection {
    hasGetterOrSetter() {
        return !!this.getSignature || !!this.setSignature;
    }
    getAllSignatures() {
        let result = [];
        if (this.signatures) {
            result = result.concat(this.signatures);
        }
        if (this.indexSignature) {
            result.push(this.indexSignature);
        }
        if (this.getSignature) {
            result.push(this.getSignature);
        }
        if (this.setSignature) {
            result.push(this.setSignature);
        }
        return result;
    }
    traverse(callback) {
        for (const parameter of lodash_1.toArray(this.typeParameters)) {
            if (callback(parameter, abstract_1.TraverseProperty.TypeParameter) === false) {
                return;
            }
        }
        if (this.type instanceof index_1.ReflectionType) {
            if (callback(this.type.declaration, abstract_1.TraverseProperty.TypeLiteral) === false) {
                return;
            }
        }
        for (const signature of lodash_1.toArray(this.signatures)) {
            if (callback(signature, abstract_1.TraverseProperty.Signatures) === false) {
                return;
            }
        }
        if (this.indexSignature) {
            if (callback(this.indexSignature, abstract_1.TraverseProperty.IndexSignature) === false) {
                return;
            }
        }
        if (this.getSignature) {
            if (callback(this.getSignature, abstract_1.TraverseProperty.GetSignature) === false) {
                return;
            }
        }
        if (this.setSignature) {
            if (callback(this.setSignature, abstract_1.TraverseProperty.SetSignature) === false) {
                return;
            }
        }
        super.traverse(callback);
    }
    toObject() {
        let result = super.toObject();
        if (this.type) {
            result.type = this.type.toObject();
        }
        if (this.defaultValue) {
            result.defaultValue = this.defaultValue;
        }
        if (this.overwrites) {
            result.overwrites = this.overwrites.toObject();
        }
        if (this.inheritedFrom) {
            result.inheritedFrom = this.inheritedFrom.toObject();
        }
        if (this.extendedTypes) {
            result.extendedTypes = this.extendedTypes.map((t) => t.toObject());
        }
        if (this.extendedBy) {
            result.extendedBy = this.extendedBy.map((t) => t.toObject());
        }
        if (this.implementedTypes) {
            result.implementedTypes = this.implementedTypes.map((t) => t.toObject());
        }
        if (this.implementedBy) {
            result.implementedBy = this.implementedBy.map((t) => t.toObject());
        }
        if (this.implementationOf) {
            result.implementationOf = this.implementationOf.toObject();
        }
        return result;
    }
    toString() {
        let result = super.toString();
        if (this.typeParameters) {
            const parameters = [];
            this.typeParameters.forEach((parameter) => {
                parameters.push(parameter.name);
            });
            result += '<' + parameters.join(', ') + '>';
        }
        if (this.type) {
            result += ':' + this.type.toString();
        }
        return result;
    }
}
exports.DeclarationReflection = DeclarationReflection;
//# sourceMappingURL=declaration.js.map