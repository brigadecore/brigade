import * as ts from 'typescript';
import { ReflectionKind, DeclarationReflection } from '../../models/index';
import { Context } from '../context';
export declare function createDeclaration(context: Context, node: ts.Declaration, kind: ReflectionKind, name?: string): DeclarationReflection | undefined;
