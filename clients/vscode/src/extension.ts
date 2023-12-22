import * as vscode from 'vscode';


import {
	Executable,
	LanguageClient,
	LanguageClientOptions,
	ServerOptions,
	TransportKind
} from 'vscode-languageclient/node';
import { getHelmLsExecutable } from './util/executable';

let client: LanguageClient;

export async function activate(_: vscode.ExtensionContext) {


	const helmLsExecutable = await getHelmLsExecutable()

	if (!helmLsExecutable) {
		vscode.window.showErrorMessage('Helm Ls executable not found');
		return;
	}

	console.log("Launching " + helmLsExecutable)

	const executable: Executable = {
		command: helmLsExecutable,
		args: [
			"serve"
		],
		transport: TransportKind.stdio
	}

	const serverOptions: ServerOptions = {
		run: executable,
		debug: executable
	};

	const clientOptions: LanguageClientOptions = {
		documentSelector: [{ scheme: 'file', language: 'helm' }],
		synchronize: {}
	};

	client = new LanguageClient(
		'helm-ls',
		'Helm Language Server',
		serverOptions,
		clientOptions
	);

	client.start();
}

export function deactivate(): Thenable<void> | undefined {
	console.log('deactivate');
	if (!client) {
		return undefined;
	}
	return client.stop();
}
