

import {Injectable, NgModuleFactory} from '@angular/core';

import {PLUGIN_EXTERNALS_MAP} from './pluginexternals';
import {PluginLoaderService} from './pluginloader.service';
import {PluginsConfigService} from '../global/plugin';

const systemJS = window.System;

@Injectable()
export class ClientPluginLoaderService extends PluginLoaderService {
  constructor(private pluginsConfigService_: PluginsConfigService) {
    super();
  }

  provideExternals() {
    Object.keys(PLUGIN_EXTERNALS_MAP).forEach(externalKey =>
      window.define(externalKey, [], () => {
        // @ts-ignore
        return PLUGIN_EXTERNALS_MAP[externalKey];
      }),
    );
  }

  load<T>(pluginName: string): Promise<NgModuleFactory<T>> {
    const plugins = this.pluginsConfigService_.pluginsMetadata();
    const plugin = plugins.find(p => p.name === pluginName);
    if (!plugin) {
      throw Error(`Can't find plugin "${pluginName}"`);
    }

    const depsPromises = (plugin.dependencies || []).map(dep => {
      const dependency = plugins.find(d => d.name === dep);
      if (!dependency) {
        throw Error(`Can't find dependency "${dep}" for plugin "${pluginName}"`);
      }

      return systemJS.import(dependency.path).then(m => {
        window['define'](dep, [], () => m.default);
      });
    });

    return Promise.all(depsPromises).then(() => {
      return systemJS.import(plugin.path).then(module => module.default.default);
    });
  }
}
